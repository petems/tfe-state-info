Feature: Version Command

  Background:
    Given I have "go" command installed
    When I run `go build -o ../../bin/tfe-state-info-int-test ../../main.go`
    Then the exit status should be 0

  Scenario:
    Given a build of tfe-state-info
    When I run `bin/tfe-state-info-int-test`
    Then the output should contain:
      """"
      NAME:
        tfe-state-info - A simple cli app to return state information from TFE
      """"
