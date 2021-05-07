package utils

import (
	"errors"
	"log"
	"os/exec"
	"regexp"
)

const SupervisorController = "/usr/bin/supervisorctl"

// Execute supervisorctl update
func UpdateSupervisor() error {
	log.Printf("Entry UpdateSupervisor: %s", SupervisorController)
	out, err := exec.Command(SupervisorController, "update").Output()
	// if there is an error with our execution
	// handle it here
	if err != nil {
		log.Fatalf("%s", err)
		return err
	}
	log.Printf("update supervisor result: %s", out)
	return nil
}

// Get supervisor service state
func SupervisorServiceState(serviceName string) (string, error) {
	log.Printf("Supervisor service name: %s\n", serviceName)
	out, err := exec.Command(SupervisorController, "status", serviceName).Output()
	if err != nil {
		log.Printf("%s", err)
		return "", err
	}
	log.Printf("Status: %s\n", out)
	pattern := regexp.MustCompile(`(?P<name>[a-zA-Z]+).*(?P<status>(STARTING|STOPPED|RUNNING)+).*`)
	if !pattern.MatchString(string(out)) {
		return "", errors.New("state not match")
	}

	matches := pattern.FindStringSubmatch(string(out))
	lastIndex := pattern.SubexpIndex("status")
	log.Printf("last => %d\n", lastIndex)
	return matches[lastIndex], nil
}

// Do supervisor service action
func DoAction(serviceName, action string) (string, error) {
	out, err := exec.Command(SupervisorController, action, serviceName).Output()
	if err != nil {
		log.Printf("%s", err)
		return "", err
	}
	log.Printf("Supervisor %s action=%s action result: %s\n", serviceName, action, out)
	return string(out), nil
}
