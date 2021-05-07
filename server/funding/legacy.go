package funding

import (
	"errors"
	"ezcoinrobot/server/utils"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/bigkevmcd/go-configparser"
)

func LegacyCreateSupervisorConfig(username, currency string) bool {
	log.Printf("Entry LegacyCreateSupervisorConfig username: %s, currency: %s", username, currency)
	configFilename := legacyRobotFilename(username, currency)
	log.Printf("supervisor config name: %s", configFilename)
	if utils.IsConfigExist(configFilename) {
		return true
	}

	config := configparser.New()
	sectionName := legacyProgramSectionName(username, currency)
	log.Printf("section name: %s", sectionName)
	if err := config.AddSection(sectionName); err != nil {
		log.Fatalln(err)
		return false
	}

	// Set program section
	sectionDict, _ := getSectionDict(username, currency)
	for key, val := range sectionDict {
		config.Set(sectionName, key, val)
	}

	err := config.SaveWithDelimiter(configFilename, "=")
	if err != nil {
		log.Fatal(err)
		return false
	}
	return true
}

func LegacyRobotState(username, currency string) (string, error) {
	supervisorStates := []string{"STARTING", "STOPPED", "RUNNING"}
	serviceName := LegacyRobotServiceName(username, currency)
	state, err := utils.SupervisorServiceState(serviceName)
	if err != nil {
		return "", err
	}

	log.Printf("Robot state %s", state)
	if !contains(supervisorStates, state) {
		return "", errors.New("Unknow state")
	}
	return state, nil
}

// Get legacy funding robot service name
func LegacyRobotServiceName(username, currency string) string {
	return strings.ToLower(username) + "_" + currency
}

func LegacyReplaceRobotCurrent(username, fromCurrency, toCurrency string) error {
	configFilename := legacyRobotFilename(username, fromCurrency)
	targetConfigFilename := legacyRobotFilename(username, toCurrency)

	// remove current config if target file exist
	if utils.IsConfigExist(targetConfigFilename) {
		log.Printf("Target %s is exist\n", targetConfigFilename)
		if utils.IsConfigExist(configFilename) {
			err := os.Remove(configFilename)
			check(err)
			return err
		}
	}

	// create target file if current file not exist
	if !utils.IsConfigExist(configFilename) {
		log.Printf("%s is not exist\n", configFilename)
		if !LegacyCreateSupervisorConfig(username, toCurrency) {
			return errors.New("create supervisor failed")
		} else {
			return nil
		}
	}

	read, err := ioutil.ReadFile(configFilename)
	check(err)

	newContents := strings.Replace(string(read), fromCurrency, toCurrency, -1)
	err = ioutil.WriteFile(configFilename, []byte(newContents), 0)
	check(err)

	err = os.Rename(configFilename, targetConfigFilename)
	check(err)

	return err
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func legacyRobotFilename(username, currency string) string {
	return supervisorConfigPath + strings.ToLower(username) + "_" + currency + ".conf"
}

func legacyProgramSectionName(username, currency string) string {
	return fmt.Sprintf("program:%username_%s", username, currency)
}
