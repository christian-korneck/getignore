/*
Copyright © 2021 Christian Korneck <christian@korneck.de>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// getignoreCmd represents the getignore command
var getignoreCmd = &cobra.Command{
	Use:     " [language ...]",
	Aliases: []string{"getignore"},
	Example: "getignore python go visualstudiocode >> .gitignore",
	Short:   "print gitignore template for a language",
	Long: `
getignore is a CLI client to GitHub's .gitignore templates.
List and print .gitingore templates for a wide variety of 
languages from the terminal. 
	
	`,
	Run: func(cmd *cobra.Command, args []string) {
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// fall back to usage screen when no args and flags specified
		if !viper.GetBool("list") && len(args) < 1 {
			cmd.Help()
			os.Exit(0)
		}

		rc := RestClient{
			Client:  http.DefaultClient,
			BaseURL: "",
		}

		output, err := rc.Run(args)

		if err != nil {
			log.Fatalf(err.Error())
		}

		if output == "" {
			log.Warn("output is empty (probably not what you wanted?")
		}

		fmt.Println(output)

		return nil
	},
}

//remove aliases from usage template
const usageTemplate = `
Usage:{{if .Runnable}}
{{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
{{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}
Examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}
Available Commands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
{{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}
Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}
Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}
Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
{{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}
Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`

const helpTemplate = `{{with (or .Long .Short)}}{{. | trimTrailingWhitespaces}}
{{end}}{{if or .Runnable .HasSubCommands}}{{.UsageString}}{{end}}`

type RestClient struct {
	Client  *http.Client
	BaseURL string
}

type TreeApiResponse struct {
	Sha  string `json:"sha"`
	Url  string `json:"url"`
	Tree []struct {
		Path     string `json:"path"`
		Mode     string `json:"mode"`
		Treetype string `json:"type"`
		Sha      string `json:"sha"`
		Url      string `json:"url"`
	} `json:"tree"`
	Truncated bool `json:"truncated"`
}

func getPaths(baseurl string) (paths []string, err error) {

	if baseurl == "" {
		baseurl = "https://api.github.com"
	}

	url := fmt.Sprintf("%s/repos/github/gitignore/git/trees/master?recursive=1", baseurl)

	resp, err := http.Get(url)
	if err != nil {
		return []string{}, fmt.Errorf("error making request to tree api: %s", err)
	}

	statusOK := resp.StatusCode >= 200 && resp.StatusCode < 300
	if !statusOK {
		return []string{}, fmt.Errorf("bad response from tree api: %s", resp.Status)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []string{}, fmt.Errorf("error reading tree api response body: %s", err)
	}

	var apires TreeApiResponse
	err = json.Unmarshal(body, &apires)
	if err != nil {
		return []string{}, fmt.Errorf("tree api response contained invalid json: %s", err)
	}

	suffix := ".gitignore"

	for _, e := range apires.Tree {

		if strings.HasSuffix(e.Path, suffix) && e.Treetype == "blob" {
			path := strings.TrimSuffix(e.Path, suffix)
			paths = append(paths, path)
		}
	}

	return paths, nil
}

func (rc *RestClient) Run(args []string) (output string, err error) {

	baseurl := ""

	if rc.BaseURL != "" {
		baseurl = rc.BaseURL
	}
	paths, err := getPaths(baseurl)

	if err != nil {
		return "", err
	}

	if viper.GetBool("list") {
		for _, path := range paths {
			output = fmt.Sprintf("%s\n%s", output, strings.ToLower(path))
		}
		return output, nil
	}

	var langs []string

	for _, lang := range args {
		found := false
		//prefer exact path
		for _, path := range paths {
			if strings.EqualFold(lang, path) {
				lang = path
				found = true
				break
			}
		}
		//if not found, fall back to the first matching suffix
		if !found {
			for _, path := range paths {
				if strings.HasSuffix(strings.ToLower(path), strings.ToLower("/"+lang)) {
					lang = path
					found = true
					break
				}
			}
		}
		//fail if not found
		if !found {
			return "", fmt.Errorf("language \"%s\" not found", lang)
		}

		langs = append(langs, lang)
	}

	if rc.BaseURL != "" {
		baseurl = rc.BaseURL
	} else {
		baseurl = "https://raw.githubusercontent.com"
	}

	for _, lang := range langs {
		url := fmt.Sprintf("%s/github/gitignore/master/%s.gitignore", baseurl, lang)

		resp, err := http.Get(url)
		if err != nil {
			return "", fmt.Errorf("error making request to content api for \"%s\": %s", lang, err)
		}

		statusOK := resp.StatusCode >= 200 && resp.StatusCode < 300
		if !statusOK {
			return "", fmt.Errorf("bad response from content api for \"%s\": %s", lang, resp.Status)
		}

		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("error reading content api response body for \"%s\": %s", lang, err)
		}

		_ = body
		output = fmt.Sprintf("%s\n\n# --- start %s --- \n\n%s\n\n# --- end %s ---", output, lang, string(body), lang)

	}

	return output, nil
}

func init() {

	rootCmd.AddCommand(getignoreCmd)
	_ = usageTemplate
	_ = helpTemplate
	getignoreCmd.Parent().SetUsageTemplate(usageTemplate)
	getignoreCmd.Parent().SetHelpTemplate(helpTemplate)

	getignoreCmd.Flags().BoolP("list", "l", false, "list available gitignore templates")
	viper.BindPFlag("list", getignoreCmd.Flags().Lookup("list"))

}
