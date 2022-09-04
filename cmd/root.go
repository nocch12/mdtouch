/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

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
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

const sep = string(os.PathSeparator)

// rootCmd represents the base command when called without any subcommands
var (
	forceFlg bool

	rootCmd = &cobra.Command{
		Use:   "mdtouch",
		Short: "Create directories and touch files at the same time.",
		Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
		// Uncomment the following line if your bare application
		// has an action associated with it:
		Run: func(cmd *cobra.Command, args []string) {
			// 引数がない場合はエラー
			if len(args) == 0 {
				fmt.Println("error: One or more arguments are required")
				return
			}

			defaultUmask := syscall.Umask(0)
			for _, file := range args {
				if err := command(file); err != nil {
					fmt.Println("error: " + err.Error())
					break
				}
			}
			syscall.Umask(defaultUmask)
		},
	}
)

func command(arg string) error {
	// ディレクトリで終わっている場合はエラー
	if strings.HasSuffix(arg, sep) {
		return errors.New("argument must be a file")
	}

	dirname, _ := filepath.Split(arg)

	// forceフラグがなければディレクトリ作成を確認する
	if !forceFlg {
		// 末尾スラッシュを削除
		dirList := strings.Split(dirname[:len(dirname)-1], sep)

		// 1階層ずつディレクトリの存在をチェック
		curDir := ""         // チェック中ディレクトリ
		dirString := ""      // 標準出力用文字列
		needsCreate := false // ディレクトリを作成するか
		for _, dir := range dirList {
			curDir += dir + sep // / ./ hoge/fuga/
			if _, err := os.Stat(curDir); os.IsNotExist(err) {
				needsCreate = true
				// 作成するディレクトリに色を付けて出力
				dirString += toRed(dir) + sep
			} else {
				dirString += dir + sep
			}
		}
		if needsCreate {
			fmt.Printf("directory: %s", dirString+"\n")
			fmt.Println("allow directory creation? [y/n]")
			scanner := bufio.NewScanner(os.Stdin)
			scanner.Scan()
			if scanner.Text() != "y" {
				fmt.Println("'y' was not entered. Skip this process.")
				return nil
			}
		}
	}

	// ディレクトリ再帰作成
	if err := os.MkdirAll(dirname, os.ModePerm); err != nil {
		return err
	}
	os.Chmod(dirname, os.ModePerm)

	// ファイル更新
	if err := touch(arg); err != nil {
		return err
	}
	return nil
}

// touchコマンド
func touch(filename string) error {
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		file, err := os.Create(filename)
		if err != nil {
			return err
		}
		defer file.Close()
		os.Chmod(filename, os.ModePerm)
	} else {
		currentTime := time.Now().Local()
		err = os.Chtimes(filename, currentTime, currentTime)
		if err != nil {
			return err
		}
	}
	return nil
}

// 出力時の文字色を赤色にする
func toRed(s string) string {
	return fmt.Sprintf("\x1b[31m%s\x1b[0m", s)
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolVarP(&forceFlg, "force", "f", false, "Forces creation of directories and touches to files")
}
