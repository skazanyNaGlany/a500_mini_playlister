// THEA500 MINI PLAYLISTER v0.1
//
// THEA500 MINI playlist generator.
//
//
// MIT License
//
// Copyright (c) 2022 skazanyNaGlany
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/thoas/go-funk"
	"github.com/yargevad/filepathx"
)

const app_name = "THEA500 MINI PLAYLISTER"
const app_version = "0.1"
const adf_pattern = "*.adf"
const m3u_pattern = "*.m3u"
const m3u_extension = "m3u"
const str_path_separator = string(os.PathSeparator)
const str_line_break = "\n"

var re_similar_rom = regexp.MustCompile(`\(Disk\ \d\ of\ \d\)`)
var directory string = ""
var keep_existing = false

func getFullAppName() string {
	return fmt.Sprintf("%v v%v", app_name, app_version)
}

func printAppName() {
	log.Println(
		getFullAppName())
	log.Println()
	log.Println("THEA500 MINI playlist generator.")
	log.Println()
}

func printAppInfo() {
	log.Println("This app will automatically generate")
	log.Println("M3U playlist files from all ADF files")
	log.Println("found in current and sub directories.")
	log.Println()
	log.Println("Add . as a prefix to the directory name")
	log.Println("to skip it.")
	log.Println()
	log.Println("Just run it and wait for finish.")
	log.Println()
	log.Println("This app will delete all your previous")
	log.Println("M3U files so use with CAUTION. Set a specific")
	log.Println("M3U file as read-only so it will not be deleted.")
	log.Println()
}

func printUsages() {
	log.Printf("Usage: %v <option>", os.Args[0])

	log.Println()
	log.Println("Options:")

	log.Println("\t-h, --help")
	log.Println("\t\t\t show this help")
	log.Println()
	log.Println("\t-d, --directory <directory>")
	log.Println("\t\t\t use <directory> instead of current app directory")
	log.Println()
	log.Println("\t-k, --keep-existing-m3u")
	log.Println("\t\t\t do not delete existing m3u files")
	log.Println()
}

func changeCurrentWorkingDir() {
	exeDir := filepath.Dir(os.Args[0])
	os.Chdir(exeDir)
}

func checkPlatform() {
}

func printUsagesExit() {
	printAppInfo()
	printUsages()

	os.Exit(1)
}

func processCommandLineArgs() {
	directory = filepath.Dir(os.Args[0])

	for i := range os.Args {
		arg_value := os.Args[i]

		if arg_value == "-h" || arg_value == "--help" {
			printUsagesExit()
		} else if arg_value == "-d" || arg_value == "--directory" {
			if i+1 == len(os.Args) {
				printUsagesExit()
			}

			i++

			directory = os.Args[i]
		} else if arg_value == "-k" || arg_value == "--keep-existing-m3u" {
			keep_existing = true
		}
	}
}

func deletePlaylists() {
	if keep_existing {
		return
	}

	log.Println("Deleting previous", m3u_pattern, "files...")
	log.Println("Searching for", m3u_pattern, "files in", directory, "...")

	m3u_files, err := filepathx.Glob(directory + m3u_pattern)

	if err != nil {
		log.Fatalln(err)
	}

	for _, ifile := range m3u_files {
		ifile_relative, err := filepath.Rel(directory, ifile)

		if err != nil {
			log.Println(err)
			continue
		}

		// skip hidden file
		if strings.HasPrefix(ifile_relative, ".") {
			continue
		}

		if !canWrite(ifile) {
			log.Println("Skipping read-only", ifile)
			continue
		}

		log.Println("Deleting", ifile)

		os.Remove(ifile)
	}
}

func fixPaths() {
	var err error

	directory = filepath.FromSlash(directory)
	directory, err = filepath.Abs(directory)

	if err != nil {
		log.Fatalln(err)
	}

	if !strings.HasSuffix(directory, str_path_separator) {
		directory += str_path_separator
	}
}

func canWrite(filepath string) bool {
	f, err := os.OpenFile(filepath, os.O_WRONLY, 0)

	if err != nil {
		if os.IsPermission(err) {
			return false
		}
	}

	f.Close()

	return true
}

func filenameSplitText(basename string) (string, string) {
	extension := filepath.Ext(basename)
	filename := strings.TrimSuffix(basename, extension)

	return filename, extension
}

func directoryExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func checkDirectoryExists() {
	if !directoryExists(directory) {
		log.Fatalln(directory, "does not exists")
	}
}

func findSimilarRoms(rom_pathname string) ([]string, string) {
	// find similar rom files, for example for rom:
	// e:\projects\a500_mini_playlister\Superfrog (1993)(Team 17)[cr CSL][a](Disk 1 of 4).adf
	//
	// similar:
	// e:\projects\a500_mini_playlister\Superfrog (1993)(Team 17)[cr CSL][a](Disk 1 of 4).adf
	// e:\projects\a500_mini_playlister\Superfrog (1993)(Team 17)[cr CSL][a](Disk 2 of 4).adf
	// e:\projects\a500_mini_playlister\Superfrog (1993)(Team 17)[cr CSL][a](Disk 3 of 4).adf
	// e:\projects\a500_mini_playlister\Superfrog (1993)(Team 17)[cr CSL][a](Disk 4 of 4).adf
	//
	// clean_filename:
	// Superfrog (1993)(Team 17)[cr CSL][a]
	dirname := filepath.Dir(rom_pathname)
	basename := filepath.Base(rom_pathname)
	similar := []string{rom_pathname}

	match := re_similar_rom.FindAll([]byte(basename), -1)

	if len(match) != 1 {
		return similar, ""
	}

	no_disc_filename := string(re_similar_rom.ReplaceAll([]byte(basename), []byte("")))
	clean_filename, extension := filenameSplitText(no_disc_filename)
	files, err := filepathx.Glob(dirname + str_path_separator + "*" + extension)

	clean_filename = strings.TrimSpace(clean_filename)

	if err != nil {
		return similar, clean_filename
	}

	sort.Strings(files)

	for _, ifile := range files {
		ifile_basename := filepath.Base(ifile)

		if strings.HasPrefix(ifile_basename, clean_filename) && strings.HasSuffix(ifile_basename, extension) {
			ifile_match := re_similar_rom.FindAll([]byte(ifile_basename), -1)
			if len(ifile_match) == 1 {
				if !funk.Contains(similar, ifile) {
					similar = append(similar, ifile)
				}
			}
		}
	}

	sort.Strings(similar)

	return similar, clean_filename
}

func createPlaylistFromFiles(pathname_m3u string, files []string) error {
	f, err := os.Create(pathname_m3u)

	if err != nil {
		return err
	}

	defer f.Close()

	for _, ifile := range files {
		// always use linux path separator
		ifile = strings.ReplaceAll(ifile, str_path_separator, "/")

		f.WriteString(ifile + str_line_break)
	}

	return nil
}

func printM3USimilar(pathname_m3u string, files []string) {
	log.Println("Creating", pathname_m3u)

	for _, ifile := range files {
		log.Println("\t", ifile)
	}
}

func filesToRelative(root_directory string, files []string) []string {
	relative_files := []string{}

	for _, ifile := range files {
		relative_pathname, err := filepath.Rel(root_directory, ifile)

		if err != nil {
			relative_files = append(relative_files, ifile)
		} else {
			relative_files = append(relative_files, relative_pathname)
		}
	}

	return relative_files
}

func createPlaylists() {
	log.Println("Generating", m3u_pattern, "files from", m3u_pattern, "files...")
	log.Println("Searching for", adf_pattern, "files in", directory, "...")

	processed_adfs := []string{}
	adf_files, err := filepathx.Glob(directory + "**" + str_path_separator + adf_pattern)

	if err != nil {
		log.Fatalln(err)
	}

	sort.Strings(adf_files)

	for _, ifile_pathname := range adf_files {
		ifile_relative, err := filepath.Rel(directory, ifile_pathname)

		if err != nil {
			log.Println(err)
			continue
		}

		// skip file which its name or directory starts from .
		// so it is hidden
		if strings.HasPrefix(ifile_relative, ".") {
			continue
		}

		if funk.Contains(processed_adfs, ifile_pathname) {
			continue
		}

		similar, clean_filename := findSimilarRoms(ifile_pathname)

		if clean_filename == "" {
			// rom file does not have (Disk n of n) in its name
			// so there will be no clean_filename
			// use first rom
			clean_filename, _ = filenameSplitText(similar[0])
		}

		clean_filename = filepath.Base(clean_filename)
		target_pathname_m3u := directory + clean_filename + "." + m3u_extension

		if fileExists(target_pathname_m3u) {
			log.Println("Skipping existing", target_pathname_m3u)
			continue
		}

		processed_adfs = append(processed_adfs, similar...)
		processed_adfs = funk.UniqString(processed_adfs)

		similar = filesToRelative(directory, similar)

		printM3USimilar(target_pathname_m3u, similar)

		err2 := createPlaylistFromFiles(target_pathname_m3u, similar)

		if err2 != nil {
			log.Println(err2)
		}
	}
}

func runApp() {
	processCommandLineArgs()
	fixPaths()
	checkDirectoryExists()
	deletePlaylists()
	createPlaylists()
}

func main() {
	printAppName()
	checkPlatform()
	changeCurrentWorkingDir()

	runApp()
}
