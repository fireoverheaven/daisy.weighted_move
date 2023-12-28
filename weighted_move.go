package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/jmcvetta/randutil"
)

var (
	BUFFERSIZE int64
	handler    slog.Handler
	logger     *slog.Logger
)

func main() {
	BUFFERSIZE = (1024 * 100)
	handler = slog.NewJSONHandler(os.Stdout, nil)
	logger = slog.New(handler)
	slog.SetDefault(logger)

	weightedRoots := parse_weightdir_file(os.Args[1])
	weighted_move(os.Args[2], weightedRoots)
}

func weighted_move(src string, weightedDests []randutil.Choice) {
	START, err := filepath.Abs(src)
	if err != nil {
		logger.Error(
			"Couldn't find absolute path to starting dir",
			slog.String("path", src),
			slog.Any("error", err),
		)
		os.Exit(1)
	}
	err = filepath.Walk(START,
		func(current string, info os.FileInfo, err error) error {
			slog_path := slog.String("src_path", current)
			if err != nil {
				logger.Error("Couldn't walk", slog_path, slog.Any("error", err))
				os.Exit(1)
			}
			logger.Info("Walking", slog_path)

			if info.IsDir() {
				logger.Info("Directory, skipping", slog_path)
				return nil
			}

			filename := filepath.Base(current)
			choice, err := randutil.WeightedChoice(weightedDests)
			if err != nil {
				logger.Error(
					"Couldn't get weighted choice",
					slog_path,
					slog.Any("weighted_dests", weightedDests),
					slog.Any("error", err),
				)
				os.Exit(1)
			}
			dest := fmt.Sprintf("%s/%s", choice.Item, filename)
			slog_path_dest := slog.String("dest_path", dest)

			logger.Info("Destination", slog_path, slog_path_dest)

			infile, err := os.Open(current)
			if err != nil {
				logger.Error(
					"Couldn't open source file",
					slog_path,
					slog_path_dest,
					slog.Any("error", err),
				)
				return nil
			}
			defer infile.Close()
			logger.Info("Source file opened", slog_path, slog_path_dest)

			outfile, err := os.Create(dest)
			if err != nil {
				logger.Warn(
					"Failed to create destination file",
					slog_path,
					slog_path_dest,
					slog.Any("error", err),
				)
				return nil
			}
			defer outfile.Close()

			fast_copy(infile, outfile)
			logger.Info("Successfully copied", slog_path, slog_path_dest)

			err = os.Remove(current)
			if err != nil {
				logger.Error(
					"Failed to delete source file after copying",
					slog_path,
					slog_path_dest,
					slog.Any("error", err),
				)
				return nil
			}
			logger.Debug("Source file deleted", slog_path, slog_path_dest)

			return nil
		})
	if err != nil {
		logger.Error(
			"Couldn't walk",
			slog.String("start_dir", START),
			slog.Any("error", err),
		)
		os.Exit(1)
	}
}

func fast_copy(fin *os.File, fout *os.File) {
	buf := make([]byte, BUFFERSIZE)
	var cum_total int64
	file_info, err := fin.Stat()
	if err != nil {
		log.Fatalf("Couldn't stat file '%s'", err)
	}
	total_size := file_info.Size()
	fmt.Printf("  { %d MB }  ", (total_size / 1024 / 1024))

	indicators := 3
	starnum := 0
	fmt.Print("[")
	for {
		n, err := fin.Read(buf)
		if err != nil && err != io.EOF {
			logger.Error(
				"fast_copy failed",
				slog.String("file_in", fin.Name()),
				slog.String("file_out", fout.Name()),
				slog.Int64("cum_total", cum_total),
				slog.Int64("total_size", total_size),
				slog.Any("error", err),
			)
			os.Exit(1)
		}
		if n == 0 {
			break
		}
		if _, err := fout.Write(buf[:n]); err != nil {
			logger.Error(
				"fast_copy failed",
				slog.String("file_in", fin.Name()),
				slog.String("file_out", fout.Name()),
				slog.Int64("cum_total", cum_total),
				slog.Int64("total_size", total_size),
				slog.Any("error", err),
			)
			os.Exit(1)
		}
		cum_total = cum_total + BUFFERSIZE
		starnum = starnum + 1

		if cum_total >= (total_size*3/4) && indicators > 0 {
			fmt.Print(" 3/4 ")
			indicators = indicators - 1
		} else if cum_total >= (total_size/2) && indicators > 1 {
			fmt.Print(" 2/4 ")
			indicators = indicators - 1
		} else if cum_total >= (total_size/4) && indicators > 2 {
			fmt.Print(" 1/4 ")
			indicators = indicators - 1
		}

		if starnum > 50 {
			fmt.Print("*")
			starnum = 0
		}
	}
	fmt.Println("]")
}

func parse_weightdir_file(weightfile_path string) []randutil.Choice {
	readFile, err := os.Open(weightfile_path)
	if err != nil {
		log.Fatalf("Couldn't open %s", weightfile_path)
	}

	fileScanner := bufio.NewScanner(readFile)
	fileScanner.Split(bufio.ScanLines)

	var weightedRoots []randutil.Choice
	for fileScanner.Scan() {
		weight, path := parse_weightdir_line(fileScanner.Text())
		root := randutil.Choice{
			Item:   path,
			Weight: weight,
		}
		weightedRoots = append(weightedRoots, root)
	}
	readFile.Close()

	return weightedRoots
}

func parse_weightdir_line(line string) (int, string) {
	info := strings.Split(line, ":")
	weight, err := strconv.Atoi(info[0])
	if err != nil {
		logger.Error(
			"Couldn't parse weightdir_file: integer conversion failed",
			slog.String("weight", info[0]),
			slog.Any("error", err),
		)
		os.Exit(1)
	}

	path, err := filepath.Abs(info[1])
	if err != nil {
		logger.Error(
			"Couldn't parse weightdir_file: couldn't get absolute path",
			slog.String("path", info[1]),
			slog.Any("error", err),
		)
		os.Exit(1)
	}
	logger.Debug(
		"Parsed a weightdir file line",
		slog.Int("weight", weight),
		slog.String("path", path),
	)

	return weight, path
}
