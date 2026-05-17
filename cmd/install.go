package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AndrewADev/bight/internal/config"
	"github.com/AndrewADev/bight/internal/hook"
	"github.com/AndrewADev/bight/internal/output"
	"github.com/spf13/cobra"
)

func installCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "install",
		Short: "Write post-checkout hook into .git/hooks/",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := hook.Install(); err != nil {
				return err
			}
			fmt.Println(output.Green("bight: hook installed"))
			return promptInitConfig()
		},
	}
}

func promptInitConfig() error {
	if _, _, err := config.Load(); err == nil {
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("bight: no config file found. Create .bight.yml? [Y/n] ")
	answer, _ := reader.ReadString('\n')
	answer = strings.TrimSpace(answer)
	if answer != "" && answer != "y" && answer != "Y" {
		return nil
	}

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	defaultProject := filepath.Base(cwd)

	fmt.Printf("  Project name [%s]: ", defaultProject)
	project, _ := reader.ReadString('\n')
	project = strings.TrimSpace(project)
	if project == "" {
		project = defaultProject
	}

	fmt.Print("  Env file path [.env]: ")
	envFile, _ := reader.ReadString('\n')
	envFile = strings.TrimSpace(envFile)
	if envFile == "" {
		envFile = ".env"
	}

	var vars []config.Var
	fmt.Print("  Add env vars to track? [Y/n] ")
	addVars, _ := reader.ReadString('\n')
	if v := strings.TrimSpace(addVars); v == "" || v == "y" || v == "Y" {
		fmt.Println("  (blank name to finish)")
		for {
			fmt.Print("    Var name: ")
			name, _ := reader.ReadString('\n')
			name = strings.TrimSpace(name)
			if name == "" {
				break
			}
			fmt.Println("    Strategy:")
			fmt.Println("      1) template  - interpolate branch/project name (default)")
			fmt.Println("      2) random    - fresh random value on each checkout")
			fmt.Print("    Choice [1]: ")
			choice, _ := reader.ReadString('\n')
			strategy := "template"
			if strings.TrimSpace(choice) == "2" {
				strategy = "random"
			}
			vars = append(vars, config.Var{Name: name, Strategy: strategy})
		}
	}

	if err := os.WriteFile(".bight.yml", []byte(config.Generate(project, envFile, vars)), 0o644); err != nil {
		return err
	}
	fmt.Println(output.Green("bight: created .bight.yml"))
	return nil
}
