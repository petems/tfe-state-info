Feature: Version Command

  Background:
    Given I have "go" command installed
    When I run `go build -o ../../bin/tfe-state-info-int-test ../../main.go`
    Then the exit status should be 0

  Scenario: Version with no flags
    Given a build of tfe-state-info
    When I run `bin/tfe-state-info-int-test --version`
    Then the output should contain:
      """""

      tfe-state-info version 0.1.0-development
      """""