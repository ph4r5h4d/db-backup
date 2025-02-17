package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

var tableFlag string
var fileName string
var filePath string
var fileType string
var configFile string
var dbUsername string
var dbPassword string
var dbDatabase string
var dbHostname string
var dbPortNumber string
var backupDate string
var bashCommand string
var includeDate bool

type LaravelEnvFile map[string]string

/**
read value from env file
*/
func ReadPropertiesFile(filename string) (LaravelEnvFile, error) {

	config := LaravelEnvFile{}

	if len(filename) == 0 {
		return config, nil
	}
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if equal := strings.Index(line, "="); equal >= 0 {
			if key := strings.TrimSpace(line[:equal]); len(key) > 0 {
				value := ""
				if len(line) > equal {
					value = strings.TrimSpace(line[equal+1:])
				}
				config[key] = value
			}
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
		return nil, err
	}

	return config, nil
}

/**
check if array contains an string value or not
*/
func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// getValidFileTypes returns an array of acceptable file extensions
func getValidFileTypes() []string {
	return []string{"sql","zip","gz"}
}

func getOutputFilePath() string {
	if includeDate {
		return filePath + fileName + backupDate
	} else {
		return filePath + fileName
	}
}

/**
parse input flags
*/
func parseFlags() {
	flag.StringVar(&tableFlag, "t", "*", "name of table(s) you want to dump comma separated.")
	flag.StringVar(&fileName, "n", "dump", "output file name")
	flag.StringVar(&filePath, "p", "", "output file path")
	flag.StringVar(&fileType, "f", "sql", "output file type (sql|zip|gz)")
	flag.BoolVar(&includeDate, "d", false, "add date to output files")

	flag.StringVar(&dbUsername, "db_user", "", "database user name")
	flag.StringVar(&dbPassword, "db_pass", "", "database user password")
	flag.StringVar(&dbHostname, "db_host", "", "database host")
	flag.StringVar(&dbDatabase, "db_name", "", "database name")
	flag.StringVar(&dbPortNumber, "db_port", "3306", "database port number")

	flag.StringVar(&configFile, "config", "", "database config file")

	flag.Parse()

	if !contains(getValidFileTypes(), fileType) {
		console("invalid file type ")

	}
}

func getTableNames() string {
	if tableFlag == "*" {
		return ""
	}
	return strings.Replace(tableFlag, ",", " ", 3)
}

/**
create bash command and store it
*/
func runDumpCommand() {
	bashCommand = "mysqldump -h" +
		dbHostname +
		" -P" + dbPortNumber +
		" -u " + dbUsername +
		" -p" + dbPassword + " " +
		dbDatabase + " " +
		getTableNames() + "  > " +
		getOutputFilePath()

	_, err := exec.Command("sh", "-c", bashCommand).Output()
	if err != nil {
		console("Error occurred!")
		os.Exit(0)
	}
}

/**
compress output file
*/
func runCompressCommand() {
	console("compressing dump file... ")
	if fileType == "zip" {
		bashCommand = "zip " + getOutputFilePath() + ".zip " + getOutputFilePath()
		_, err := exec.Command("sh", "-c", bashCommand).Output()
		if err != nil {
			console("Cannot compress file.")
			os.Exit(0)
		}
	} else {
		bashCommand = "gzip -c " + getOutputFilePath() + " > " + getOutputFilePath() + ".gz"
		_, err := exec.Command("sh", "-c", bashCommand).Output()
		if err != nil {
			console("Cannot compress file.")
			os.Exit(0)
		}
	}
}

/**
read environment file
*/
func readEnv() {
	var config, _ = ReadPropertiesFile(configFile)
	dbDatabase = config["DB_DATABASE"]
	dbHostname = config["DB_HOST"]
	dbUsername = config["DB_USERNAME"]
	dbPassword = config["DB_PASSWORD"]
	dbPortNumber = config["DB_PORT"]
}

/**
get iso Date & time
*/
func isoDate(kebabcase bool) string {
	dt := time.Now()
	if kebabcase {
		var kc = strings.Replace(dt.Format("01-02-2006 15:04:05"), " ", "-", -1)
		kc = strings.Replace(kc, ":", "-", -1)
		return kc
	}
	return dt.Format("01-02-2006 15:04:05")

}

/**
console output
*/
func console(message string) {
	fmt.Println(isoDate(false) + " :: " + message)
}

func main() {
	backupDate = isoDate(true)
	parseFlags()
	if configFile != "" {
		readEnv()
	}
	console("dumping database...")
	runDumpCommand()
	if fileType == "zip" || fileType == "gz" {
		runCompressCommand()
	}
	console("done!")
}
