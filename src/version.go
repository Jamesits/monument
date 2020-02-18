package main

import "fmt"

var versionMajor = 0
var versionMinor = 0
var versionRevision = 1
var versionGitCommitHash string
var versionCompileTime string
var versionCompileHost string
var versionGitStatus string

func getVersionNumberString() string {
	return fmt.Sprintf("%d.%d.%d", versionMajor, versionMinor, versionRevision)
}

func getVersionFullString() string {
	if len(versionCompileHost) == 0 {
		versionCompileHost = "localhost"
	}

	if len(versionGitCommitHash) == 0 {
		versionGitCommitHash = "UNKNOWN"
	}

	if len(versionCompileTime) == 0 {
		versionCompileTime = "UNKNOWN TIME"
	}

	if len(versionGitStatus) == 0 {
		versionGitStatus = "dirty"
	}

	return fmt.Sprintf("monument %s (https://github.com/Jamesits/monument, compiled on %s for commit %s (%s) at %s)", getVersionNumberString(), versionCompileHost, versionGitCommitHash, versionGitStatus, versionCompileTime)
}