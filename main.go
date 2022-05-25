package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var (
	GitFullPath = os.Getenv("GitFullPath")
	ZipFullPath = os.Getenv("ZipFullPath")

	sourceDirFullPath  string
	targetRepoNamePart string
	gitUserName        string
	gitPassword        string
)

func main() {
	flag.StringVar(&sourceDirFullPath, "source", "", "Absolute path of source directory")
	flag.StringVar(&targetRepoNamePart, "target", "", "Target repository name part")
	flag.StringVar(&gitUserName, "mode", "username", "Git user name")
	flag.StringVar(&gitPassword, "mode", "password", "Git password")
	flag.Parse()

	if len(sourceDirFullPath) < 1 || len(targetRepoNamePart) < 1 || len(gitUserName) < 1 || len(gitPassword) < 1 {
		log.Fatal("Please specify all arguments")
	}

	wd, _ := os.Getwd()
	err := Command(wd, [][]string{
		{
			ZipFullPath,
			"a",
			fmt.Sprintf("%s.zip", sourceDirFullPath),
			sourceDirFullPath,
			"-v100m",
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("zip DONE")

	err = Command(wd, [][]string{
		{
			GitFullPath,
			"clone",
			fmt.Sprintf("http://%s:%s@github.com/%s/%s.git", gitUserName, gitPassword, gitUserName, targetRepoNamePart),
		},
	})

	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("git clone DONE")

	for i := 1; ; i++ {
		filename := fmt.Sprintf("%s.zip.%03d", sourceDirFullPath, i)
		exists := Exists(filename)
		if !exists {
			break
		}

		fmt.Println(filename, "exists")

		// 20 files a time
		CopyFile(filename, filepath.Join(wd, targetRepoNamePart, GetFileNameWithExtension(filename)))
	}

	err = Command(filepath.Join(wd, targetRepoNamePart), [][]string{
		{
			GitFullPath,
			"add",
			"*",
		},
		{
			GitFullPath,
			"commit",
			"-m",
			"- upload files",
		},
		{
			GitFullPath,
			"branch",
			"-M",
			"main",
		},
		{
			GitFullPath,
			"push",
			"-u",
			"origin",
			"main",
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("DONE")
}

func Exists(filename string) bool {
	file, err := os.Open(filename)
	if errors.Is(err, os.ErrNotExist) {
		return false
	}
	defer file.Close()
	return true
}

func Command(workingDir string, commands [][]string) error {
	for _, command := range commands {
		execCmd := exec.Command(command[0], command[1:]...)
		if len(workingDir) > 0 {
			execCmd.Dir = workingDir
		}
		_, err := execCmd.CombinedOutput()
		if err != nil {
			return err
		}
	}
	return nil
}

func CopyFile(srcFileFullPath, destFileFullPath string) error {
	bytesRead, err := ioutil.ReadFile(srcFileFullPath)

	if err != nil {
		log.Fatal(err)
	}

	return ioutil.WriteFile(destFileFullPath, bytesRead, 0644)
}

func GetFileNameWithExtension(path string) string {
	return path[strings.LastIndex(path, "\\")+1:]
}
