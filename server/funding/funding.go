package funding

import (
	"errors"
	"ezcoinrobot/server/utils"
	"fmt"
	"log"

	"strings"

	"github.com/bigkevmcd/go-configparser"
)

const (
	supervisorConfigPath = "/etc/supervisor/conf.d/"
	pythonPath           = "/home/john/SuperFundingBot/.venv/bin/python"
	createFundingProgram = "/home/john/bitfinex-funding-robot/create_funding_offers3.py"
)

// Create supervisor funding config
func CreateSupervisorFunding(username string, currency string) error {
	return createFundingConfig(username, currency)
}

func FundingRobotState(username, currency string) (string, error) {
	supervisorStates := []string{"STARTING", "STOPPED", "RUNNING"}
	state, err := robotState(username, currency)
	if err != nil {
		return "", err
	}

	log.Printf("Robot state %s", state)
	if !contains(supervisorStates, state) {
		return "", errors.New("Unknow state")
	}
	return state, nil
}

// Restart funding supervisor service
func RestartFundingRobot(username, currency string) (string, error) {
	return doFundingRobotAction(username, currency, "restart")
}

// Stop funding supervisor service
func StopFundingRobot(username, currency string) (string, error) {
	return doFundingRobotAction(username, currency, "stop")
}

// Start funding supervisor service
func StartFundingRobot(username, currency string) (string, error) {
	return doFundingRobotAction(username, currency, "start")
}

func MigrateFundingRobotServiceName(username, fromCurrency, toCurrency string) error {
	return replaceConfigSection(username, fromCurrency, toCurrency)
}

// Get funding robot service name
func FundingRobotServiceName(username, currency string) string {
	return fmt.Sprintf("%s:%s", robotGroupName(username), currency)
}

func robotFilename(username string) string {
	return supervisorConfigPath + strings.ToLower(username) + "_funding.conf"
}

func robotGroupName(username string) string {
	return fmt.Sprintf("%s_FUNDING", strings.ToLower(username))
}

func robotServiceName(username, currency string) string {
	return fmt.Sprintf("%s:%s", robotGroupName(username), currency)
}

func groupSectionName(username string) string {
	return fmt.Sprintf("group:%s", robotGroupName(username))
}

func programSectionName(currency string) string {
	return fmt.Sprintf("program:%s", currency)
}

func createFundingConfig(username, currency string) error {
	log.Printf("Entry createFundingConfig username: %s, currency: %s", username, currency)
	configFilename := robotFilename(username)
	log.Printf("supervisor config name: %s", configFilename)
	if utils.IsConfigExist(configFilename) {
		return errors.New("config file exist, can't create again.")
	}

	config := configparser.New()
	groupSectionName := groupSectionName(username)
	log.Printf("group section name: %s", groupSectionName)
	if err := config.AddSection(groupSectionName); err != nil {
		log.Fatalln(err)
		return err
	}

	sectionName := fmt.Sprintf("program:%s", currency)
	log.Printf("section name: %s", sectionName)
	if err := config.AddSection(sectionName); err != nil {
		log.Fatalln(err)
		return err
	}

	config.Set(groupSectionName, "programs", currency)
	// Set program section
	sectionDict, _ := getSectionDict(username, currency)
	for key, val := range sectionDict {
		config.Set(sectionName, key, val)
	}

	err := config.SaveWithDelimiter(configFilename, "=")
	if err != nil {
		log.Fatal(err)
		return err
	}
	return nil
}

func replaceConfigSection(username, fromCurrency, toCurrency string) error {
	config, _ := configparser.Parse(robotFilename(username))
	groupSectionName := groupSectionName(username)
	if !config.HasSection(groupSectionName) {
		return errors.New("There are no section found.")
	}
	// Get group programs and replace from current to new one
	groupProgramStr, err := config.Get(groupSectionName, "programs")
	if err != nil {
		log.Fatalln("group section not found")
		return err
	}
	log.Printf("current group programs %s", groupProgramStr)
	groupPrograms := strings.Split(groupProgramStr, ",")
	groupPrograms, _ = remove(groupPrograms, fromCurrency)
	groupPrograms = append(groupPrograms, toCurrency)
	log.Printf("new group programs: %s", groupPrograms)

	// Remove current section
	sectionName := fmt.Sprintf("program:%s", fromCurrency)
	log.Printf("delete section %s", sectionName)
	if !config.HasSection(sectionName) {
		return errors.New(fmt.Sprintf("There are no section %s found.", sectionName))
	}
	if err := config.RemoveSection(sectionName); err != nil {
		return err
	}

	newSectionName := fmt.Sprintf("program:%s", toCurrency)
	log.Printf("new session name %s", newSectionName)
	if err := config.AddSection(newSectionName); err != nil {
		return nil
	}
	pMap, _ := getSectionDict(username, toCurrency)
	for k, v := range pMap {
		config.Set(newSectionName, k, v)
	}
	config.Set(groupSectionName, "programs", strings.Join(groupPrograms, ","))

	err = config.SaveWithDelimiter(robotFilename(username), "=")
	if err != nil {
		log.Fatal(err)
		return err
	}
	return nil
}
func doFundingRobotAction(username, currency, action string) (string, error) {
	serviceName := robotServiceName(username, currency)
	return utils.DoAction(serviceName, action)
}

func robotState(username, currency string) (string, error) {
	serviceName := robotServiceName(username, currency)
	return utils.SupervisorServiceState(serviceName)
}

func contains(slice []string, item string) bool {
	set := make(map[string]struct{}, len(slice))
	for _, s := range slice {
		set[s] = struct{}{}
	}

	_, ok := set[item]
	return ok
}

func remove(s []string, r string) ([]string, error) {
	for i, v := range s {
		if v == r {
			return append(s[:i], s[i+1:]...), nil
		}
	}
	return s, errors.New("not found")
}

func getSectionDict(username, currency string) (map[string]string, error) {
	m := make(map[string]string)
	m["command"] = fmt.Sprintf("%s %s %s -s %s", pythonPath, createFundingProgram, username, currency)
	m["autostart"] = "true"
	m["autorestart"] = "true"
	m["user"] = "john"
	m["stderr_logfile"] = fmt.Sprintf("/var/log/ezcoin/%s_%s.err.log", username, currency)
	m["stdout_logfile"] = fmt.Sprintf("/var/log/ezcoin/%s_%s.out.log", username, currency)

	return m, nil
}
