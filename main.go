/*
 * Minio S3Verify Library for Amazon S3 Compatible Cloud Storage (C) 2016 Minio, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"fmt"
	"sort"

	"github.com/minio/cli"
	"github.com/minio/mc/pkg/console"
)

// Global number of tests implemented currently
var totalTests = 5

// Global scanBar for all tests to access and update.
var scanBar = scanBarFactory()

// Global command line flags.
var (
	s3verifyFlags = []cli.Flag{
		cli.BoolFlag{
			Name:  "help, h",
			Usage: "Show help.",
		},
	}
)

// Custom help template.
// Revert to API not command.
var s3verifyHelpTemplate = `NAME:
	{{.Name}} - {{.Usage}}

USAGE:
	{{.Name}} [COMMAND...] {{if .Flags}}[FLAGS] {{end}}
	
COMMANDS:
	{{range .Commands}}{{join .Names ", "}}{{ "\t" }}{{.Usage}}
	{{end}}{{if .Flags}}
GLOBAL FLAGS:
	{{range .Flags}}{{.}}
	{{end}}{{end}}
EXAMPLES:
	1. Run all tests on Minio server. Note play.minio.io is a public test server. You are free to
	use these secret and Access keys in all your tests.
		$ S3_URL=https://play.minio.io:9000 S3_ACCESS=Q3AM3UQ867SPQQA43P2F S3_SECRET=zuf+tfteSlswRu7BJ86wekitnifILbZam1KYY3TG s3verify 
	2. Run only makebucket and listbuckets tests on Amazon S3 server using flags.
	Note that passing access and secret keys as flags should be avoided on a multi-user server for security reasons.
		$ s3verify --access YOUR_ACCESS_KEY --secret YOUR_SECRET_KEY --url https://s3.amazonaws.com makebucket listbuckets
`

func commandNotFound(ctx *cli.Context, command string) {
	msg := fmt.Sprintf("'%s' is not a s3verify command. See 's3verify --help'.", command)
	closestCommands := findClosestCommands(command)
	if len(closestCommands) > 0 {
		msg += fmt.Sprintf("\n\nDid you mean one of these?\n")
		if len(closestCommands) == 1 {
			cmd := closestCommands[0]
			msg += fmt.Sprintf("\t'%s'\n", cmd)
		} else {
			for _, cmd := range closestCommands {
				msg += fmt.Sprintf("\t'%s'\n", cmd)
			}
		}
	}
	console.PrintC(msg)
}

func findClosestCommands(command string) []string {
	var closestCommands []string
	for _, value := range commandsTree.PrefixMatch(command) {
		closestCommands = append(closestCommands, value.(string))
	}
	sort.Strings(closestCommands)
	for _, value := range commandsTree.walk(commandsTree.root) {
		if sort.SearchStrings(closestCommands, value.(string)) < len(closestCommands) {
			continue
		}
		// 2 is arbitrary and represents the max allowed number of typed errors.
		if DamerauLevenshteinDistance(command, value.(string)) < 2 {
			closestCommands = append(closestCommands, value.(string))
		}
	}
	return closestCommands
}

// registerApp - Create a new s3verify app.
func registerApp() *cli.App {
	registerCmd(getObjectCmd)    // Register 'getobject' as a command.
	registerCmd(listBucketsCmd)  // Register 'listbuckets' as a command.
	registerCmd(removeBucketCmd) // Register 'removebucket' as a command.
	registerCmd(makeBucketCmd)   // Register 'makebucket' as a command.

	app := cli.NewApp()
	app.Usage = "Test for Amazon S3 v4 API compatibility."
	app.Author = "Minio.io"
	app.Commands = commands
	app.Name = "s3verify"
	app.Flags = append(s3verifyFlags, globalFlags...)
	app.CustomAppHelpTemplate = s3verifyHelpTemplate // Custom help template defined above.
	app.CommandNotFound = commandNotFound            // Custom commandNotFound function defined above.
	app.Action = callAllAPIs                         // Command to run if no commands are explicitly passed.
	return app
}

// callAllAPIS parse context extract flags and then call all.
// runAllCommands - Run all the available tests.
func callAllAPIs(ctx *cli.Context) {
	if ctx.GlobalString("access") != "" && ctx.GlobalString("secret") != "" && ctx.GlobalString("url") != "" { // Necessary variables passed, run all tests.
		mainListBuckets(ctx)
		mainMakeBucket(ctx)
		mainGetObject(ctx)
		mainRemoveBucket(ctx)
	} else {
		cli.ShowAppHelp(ctx)
	}
}

// main - Set up and run the app.
func main() {
	app := registerApp()
	app.RunAndExitOnError()
}
