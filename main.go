package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"

	// helpers

	bytefmt "code.cloudfoundry.org/bytefmt"
	"github.com/hashicorp/go-tfe"
	"github.com/hokaccha/go-prettyjson"
	"github.com/kataras/tablewriter"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

// Version is what is returned by the `-v` flag
const Version = "0.1.0"

// gitCommit is the gitcommit its built from
var gitCommit = "development"

func main() {
	//nolint:dupl // CLI config is repetitive and flags as duplicates
	app := &cli.App{
		Name:    "tfe-state-info",
		Usage:   "A simple cli app to return state information from TFE",
		Version: Version + "-" + gitCommit,
		Commands: []*cli.Command{
			{
				Name:  "list-workspaces",
				Usage: "List all workspaces for an Organization",
				Action: func(c *cli.Context) error {
					err := cmdTFEListWorkspaces(c)
					return err
				},
			},
			{
				Name:  "latest-statefile-size",
				Usage: "Get latest statefile size for all workspaces",
				Action: func(c *cli.Context) error {
					err := cmdTFELatestStatefileSizes(c)
					return err
				},
			},
			{
				Name:  "all-statefiles-size",
				Usage: "Get total size of all statefiles of all workspaces",
				Action: func(c *cli.Context) error {
					err := cmdTFEAllStatefileSizes(c)
					return err
				},
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "cleanup",
						Value: true,
						Usage: "Cleanup downloaded statefiles after completion",
					},
					&cli.BoolFlag{
						Name:  "totmpdir",
						Value: true,
						Usage: "Download statefiles to a temporary location instead of cwd",
					},
				},
			},
			{
				Name:  "validate",
				Usage: "Validate your current credentials",
				Action: func(c *cli.Context) error {
					err := cmdTFEValidate(c)
					return err
				},
			},
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "format",
				Value: "pretty_json",
				Usage: "The format you want them returned in, valid values are: table, json, pretty_json",
			},
			&cli.BoolFlag{
				Name:  "silent",
				Value: true,
				Usage: "Do not output anything other than errors or returned data",
			},
			&cli.BoolFlag{
				Name:  "debug",
				Value: false,
				Usage: "Show debug information, with full http logs",
			},
		},
	}

	cli.AppHelpTemplate = `NAME:
	{{.Name}} - {{.Usage}}
USAGE:
	{{.HelpName}} {{if .VisibleFlags}}[global options]{{end}}{{if .Commands}} command [command options]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{else}}[arguments...]{{end}}
	{{if len .Authors}}
AUTHOR:
	{{range .Authors}}{{ . }}{{end}}
	{{end}}{{if .Commands}}
COMMANDS:
{{range .Commands}}{{if not .HideHelp}}   {{join .Names ", "}}{{ "\t"}}{{.Usage}}{{ "\n" }}{{end}}{{end}}
TFE CONFIGURATION:
	TFE configuration is set by the common Vault environmental variables: 
		TFE_HOSTNAME: The address for the TFE server (Required)
		TFE_TOKEN: The token for the TFE server (Required) 
		TFE_ORG_NAME: The org you want to use

GLOBAL OPTIONS:
	{{range .VisibleFlags}}{{.}}
	{{end}}{{end}}{{if .Copyright }}
COPYRIGHT:
	{{.Copyright}}
	{{end}}{{if .Version}}
VERSION:
	{{.Version}}
	{{end}}
 `

	err := app.Run(os.Args)

	if err != nil {
		log.Fatal(err)
	}
}

func setupLogging(debug bool) {
	log.SetOutput(os.Stderr)
	textFormatter := new(prefixed.TextFormatter)
	textFormatter.FullTimestamp = true
	textFormatter.TimestampFormat = "2006-Jan-02 15:04:05"
	log.SetFormatter(textFormatter)
	log.SetLevel(log.FatalLevel)

	if debug {
		log.SetLevel(log.InfoLevel)
		log.Info("--debug setting detected - Info level logs enabled")
	}
}

func getENV(value string) (string, error) {
	envValue := os.Getenv(value)
	if len(envValue) == 0 {
		return "", fmt.Errorf("No ENV value for %s", value)
	}
	return envValue, nil
}

func printResults(format string, workspaceList *tfe.WorkspaceList) error {

	//nolint:dupl // JSON case gets flagged here
	switch format {
	case "json":
		certAsMarshall, err := json.Marshal(workspaceList)
		if err != nil {
			return err
		}
		fmt.Println(string(certAsMarshall))
	case "pretty_json":
		s, err := prettyjson.Marshal(workspaceList)
		if err != nil {
			return err
		}
		fmt.Println(string(s))
	case "table":
		tablePrint(workspaceList)
	default:
		return fmt.Errorf("invalid format given. valid formats: json, pretty_json, table, got: %s", format)
	}

	return nil

}

func tablePrint(tfeWorkspaceList *tfe.WorkspaceList) {

	workspaceArray := tfeWorkspaceList.Items

	data := [][]string{}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetAutoWrapText(false)
	table.SetAutoFormatHeaders(true)
	table.SetHeader([]string{"Name", "VCS Repo"})
	table.SetAlignment(tablewriter.ALIGN_CENTER)

	for _, workspace := range workspaceArray {

		vcsRepo := ""

		if workspace.VCSRepo == nil {
			vcsRepo = "<NONE>"
		} else {
			vcsRepo = workspace.VCSRepo.Identifier
		}

		data = append(data, []string{workspace.Name, vcsRepo})
	}

	for _, v := range data {
		table.Append(v)
	}

	table.Render()
}

func cmdTFEValidate(ctx *cli.Context) (err error) {

	tfeAddr, err := getENV("TFE_HOSTNAME")

	if err != nil {
		return err
	}

	tfeToken, err := getENV("TFE_TOKEN")

	if err != nil {
		return err
	}

	url, err := url.ParseRequestURI(fmt.Sprintf("https://%s", tfeAddr))
	if err != nil {
		return err
	}

	tfeConfig := &tfe.Config{
		Address:  url.String(),
		BasePath: "",
		Token:    tfeToken,
	}

	client, err := tfe.NewClient(tfeConfig)
	if err != nil {
		return err
	}

	// Create a context
	backgroundCtx := context.Background()

	user, err := client.Users.ReadCurrent(backgroundCtx)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("API Token valid - User is:", user.Username)
	}

	return nil

}

func cmdTFEListWorkspaces(ctx *cli.Context) (err error) {

	tfeAddr, err := getENV("TFE_HOSTNAME")

	if err != nil {
		return err
	}

	tfeToken, err := getENV("TFE_TOKEN")

	if err != nil {
		return err
	}

	url, err := url.ParseRequestURI(fmt.Sprintf("https://%s", tfeAddr))
	if err != nil {
		return err
	}

	tfeOrg, err := getENV("TFE_ORG_NAME")

	if err != nil {
		return err
	}

	tfeConfig := &tfe.Config{
		Address: url.String(),
		Token:   tfeToken,
	}

	client, err := tfe.NewClient(tfeConfig)
	if err != nil {
		return err
	}

	// Create a context
	backgroundCtx := context.Background()

	workspaces, err := client.Workspaces.List(backgroundCtx, tfeOrg, tfe.WorkspaceListOptions{})

	if err != nil {
		return err
	}

	err = printResults(ctx.String("format"), workspaces)

	if err != nil {
		return err
	}

	return nil

}

func cmdTFELatestStatefileSizes(ctx *cli.Context) (err error) {

	tfeAddr, err := getENV("TFE_HOSTNAME")

	if err != nil {
		return err
	}

	tfeToken, err := getENV("TFE_TOKEN")

	if err != nil {
		return err
	}

	url, err := url.ParseRequestURI(fmt.Sprintf("https://%s", tfeAddr))
	if err != nil {
		return err
	}

	tfeOrg, err := getENV("TFE_ORG_NAME")

	if err != nil {
		return err
	}

	tfeConfig := &tfe.Config{
		Address: url.String(),
		Token:   tfeToken,
	}

	client, err := tfe.NewClient(tfeConfig)
	if err != nil {
		return err
	}

	// Create a context
	backgroundCtx := context.Background()

	workspaces, err := client.Workspaces.List(backgroundCtx, tfeOrg, tfe.WorkspaceListOptions{})

	if err != nil {
		return err
	}

	listWorkspaces := workspaces.Items

	for _, workspace := range listWorkspaces {

		currentStateFile, err := client.StateVersions.Current(backgroundCtx, workspace.ID)

		if err == nil {

			filename := fmt.Sprintf("%s-latest-state-file.json", workspace.Name)

			err := downloadFile(filename, currentStateFile.DownloadURL)

			if err != nil {
				return err
			}

			fi, err := os.Stat(filename)

			if err != nil {
				return err
			}

			fileSize := bytefmt.ByteSize(uint64(fi.Size()))

			fmt.Printf("File size for %s was %v\n", workspace.Name, fileSize)
		}

	}

	if err != nil {
		return err
	}

	return nil

}

func cmdTFEAllStatefileSizes(ctx *cli.Context) (err error) {

	setupLogging(ctx.Bool("debug"))

	doCleanup := ctx.Bool("cleanup")

	if doCleanup {
		log.Info("--cleanup enabled, so statefiles will be deleted after completion")
	}

	doTmpDir := ctx.Bool("totmpdir")

	if doCleanup {
		log.Info(fmt.Sprintf("--totmpdir enabled, so statefiles will be downloaded to default tmpdir (%s)", os.TempDir()))
	}

	tfeAddr, err := getENV("TFE_HOSTNAME")

	if err != nil {
		return err
	}

	tfeToken, err := getENV("TFE_TOKEN")

	if err != nil {
		return err
	}

	url, err := url.ParseRequestURI(fmt.Sprintf("https://%s", tfeAddr))
	if err != nil {
		return err
	}

	tfeOrg, err := getENV("TFE_ORG_NAME")

	if err != nil {
		return err
	}

	tfeConfig := &tfe.Config{
		Address: url.String(),
		Token:   tfeToken,
	}

	client, err := tfe.NewClient(tfeConfig)
	if err != nil {
		return err
	}

	// Create a context
	backgroundCtx := context.Background()

	workspacesPagedList, err := client.Workspaces.List(backgroundCtx, tfeOrg, tfe.WorkspaceListOptions{
		ListOptions: tfe.ListOptions{PageSize: 20},
	})

	if err != nil {
		return err
	}

	workspaceNamesList := []string{}

	for i := 1; i <= workspacesPagedList.Pagination.TotalPages; i++ {
		pageNumber := i
		workspaceNamesFromPage, err := getWorkspacesListPage(backgroundCtx, pageNumber, client, tfeOrg)
		if err != nil {
			return err
		}
		workspaceNamesList = append(workspaceNamesList, workspaceNamesFromPage...)
	}

	if err != nil {
		return err
	}

	for _, workspace := range workspaceNamesList {

		stateFilePagedList, err := client.StateVersions.List(backgroundCtx, tfe.StateVersionListOptions{
			ListOptions: tfe.ListOptions{
				PageSize: 20,
			},
			Organization: &tfeOrg,
			Workspace:    &workspace,
		})

		statefileURLList := []string{}

		for i := 1; i <= stateFilePagedList.Pagination.TotalPages; i++ {
			pageNumber := i
			statefileURLsFromPage, err := getStatefileListPage(backgroundCtx, pageNumber, client, tfeOrg, workspace)
			if err != nil {
				return err
			}
			statefileURLList = append(statefileURLList, statefileURLsFromPage...)
		}

		if err != nil {
			return err
		}

		if err == nil {

			totalSize := 0

			for index, statefileURL := range statefileURLList {

				filename := fmt.Sprintf("%s-latest-state-file-%v.json", workspace, index)

				if doTmpDir {

					tmpDir, err := ioutil.TempDir(os.TempDir(), "tfe-state-info")

					if err != nil {
						return err
					}

					filename = fmt.Sprintf("%s/%s", tmpDir, filename)
				}

				err = downloadFile(filename, statefileURL)

				if err != nil {
					return err
				}

				log.Info(fmt.Sprintf("Downloaded to " + filename))

				fi, err := os.Stat(filename)

				if err != nil {
					return err
				}

				totalSize += int(fi.Size())

				if doCleanup {
					err = os.Remove(filename)

					if err != nil {
						return err
					}
				}

			}

			fileSize := bytefmt.ByteSize(uint64(totalSize))

			fmt.Printf("Total of all state file sizes for %s was %v (Statefile Count: %v)\n", workspace, fileSize, len(statefileURLList))
		}

	}

	return nil

}

func downloadFile(filepath string, url string) error {

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}

func getWorkspacesListPage(backgroundCtx context.Context, pageNumber int, client *tfe.Client, orgName string) ([]string, error) {

	workspaceNamesSlice := []string{}

	opts := tfe.WorkspaceListOptions{
		ListOptions: tfe.ListOptions{
			PageSize:   100,
			PageNumber: pageNumber,
		},
	}
	list, err := client.Workspaces.List(backgroundCtx, orgName, opts)
	if err != nil {
		return nil, err
	}
	for itr, ws := range list.Items {
		log.Info(fmt.Sprintf("Workspace %s - Workspace Iterate - %v", ws.Name, itr))
		workspaceNamesSlice = append(workspaceNamesSlice, ws.Name)
	}

	return workspaceNamesSlice, nil
}

func getStatefileListPage(backgroundCtx context.Context, pageNumber int, client *tfe.Client, orgName string, workspaceName string) ([]string, error) {

	stateFileURLs := []string{}

	opts := tfe.StateVersionListOptions{
		ListOptions: tfe.ListOptions{
			PageSize:   100,
			PageNumber: pageNumber,
		},
		Organization: &orgName,
		Workspace:    &workspaceName,
	}
	list, err := client.StateVersions.List(backgroundCtx, opts)
	if err != nil {
		return nil, err
	}
	for itr, state := range list.Items {
		log.Info(fmt.Sprintf("Workspace %s - Statefile Iterate - %v", workspaceName, itr))
		stateFileURLs = append(stateFileURLs, state.DownloadURL)
	}

	return stateFileURLs, nil
}
