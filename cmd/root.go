package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"github.com/KacperMalachowski/image-detector-action/pkg/strategy"
	"github.com/bmatcuk/doublestar/v4"
	set "github.com/deckarep/golang-set/v2"
	"github.com/sethvargo/go-githubactions"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger interface {
	Debugw(msg string, keysAndValues ...interface{})
	Infow(msg string, keysAndValues ...interface{})
	Errorw(msg string, keysAndValues ...interface{})
}

// Detector is responsible for detecting image URLs in files.
type Detector interface {
	Detect(file io.Reader) ([]string, error)
	IsSupported(file string) bool
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "image-detector",
	Short: "A CLI tool to detect image URLs in files",
	Long: `Image Detector is a command-line tool that scans files for image URLs.

	It supports various file types and can be extended to include more detectors.`,
	RunE: run,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringP("check-directory", "d", ".", "Directory to check for files")
	viper.BindPFlag("check-directory", rootCmd.PersistentFlags().Lookup("check-directory"))

	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose output")
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))

	rootCmd.PersistentFlags().StringP("config", "c", "", "Path to configuration file")
	viper.BindPFlag("config", rootCmd.PersistentFlags().Lookup("config"))

	rootCmd.PersistentFlags().StringSliceP("exclude", "e", []string{}, "Comma-separated list of file patterns to exclude from detection")
	viper.BindPFlag("exclude", rootCmd.PersistentFlags().Lookup("exclude"))

	viper.SetDefault("check-directory", ".")
	viper.SetDefault("verbose", false)
	viper.SetDefault("config", "")
	viper.SetDefault("exclude", []string{"**.git**", "**.git/**", "**node_modules**", "**vendor**"})

	// Read configuration from file if specified
	if configFile := viper.GetString("config"); configFile != "" {
		viper.SetConfigFile(configFile)

		if err := viper.ReadInConfig(); err != nil {
			log.Fatalf("Error reading config file: %v", err)
		}
	}
}

func run(cmd *cobra.Command, args []string) error {
	// Set up logging
	logger := createLogger().Sugar()

	logger.Info("Starting image detector...")

	availableDetectors := []Detector{
		&strategy.Generic{}, // Add your detectors here
	}

	logger.Debugw("Available detectors", "count", len(availableDetectors))

	// root directory to check
	checkDirectory := viper.GetString("check-directory")

	// Exclude patterns from command line flags
	excludes := viper.GetStringSlice("exclude")

	logger.Debugw("Excluding patterns", "patterns", excludes)

	logger.Infow("Scanning directory", "directory", viper.GetString("check-directory"))

	images, err := findImagesInFiles(
		checkDirectory,
		availableDetectors,
		excludes,
	)
	if err != nil {
		logger.Errorw("Error finding images in files", "error", err)
		os.Exit(1)
	}

	logger.Infow("Image detection completed", "imagesFound", len(images))

	err = setImagesOutput(images)
	if err != nil {
		logger.Errorw("Error setting images output", "error", err)
		os.Exit(1)
	}

	return nil
}

func setImagesOutput(images []string) error {
	output, err := json.Marshal(images)
	if err != nil {
		return fmt.Errorf("failed to marshal images: %w", err)
	}

	githubactions.SetOutput("images", string(output))

	return nil
}

// findImagesInFiles walks through the specified filesystem tree
// and use implemented detectors to find images in files.
// It runs only one detector per file, the first one that supports the file type.
// Uses generic detector for all files that are not supported by any other detector.
// It returns a slice of detected image URLs.
func findImagesInFiles(root string, detectors []Detector, excludePatterns []string) ([]string, error) {
	images := set.NewSet[string]()

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("error accessing path %q: %w", path, err)
		}

		if d.IsDir() {
			return nil // Skip directories
		}

		// Check if the file matches any exclude pattern
		for _, pattern := range excludePatterns {
			matched, err := doublestar.Match(pattern, path)
			if err != nil {
				return fmt.Errorf("error matching exclude pattern %q against path %q: %w",
					pattern, path, err)
			}
			if matched {
				return nil // Skip excluded files
			}
		}

		// Check if the detector supports the file
		for _, detector := range detectors {
			if detector.IsSupported(path) {
				file, err := os.Open(path)
				if err != nil {
					return fmt.Errorf("error opening file %q: %w", path, err)
				}
				defer file.Close()

				detectedImages, err := detector.Detect(file)
				if err != nil {
					return fmt.Errorf("error detecting images in file %q: %w", path, err)
				}

				for _, image := range detectedImages {
					images.Add(image)
				}
				return nil // Stop after the first matching detector
			}
		}

		return nil
	})

	if err != nil {
		return []string{}, fmt.Errorf("error walking directory: %w", err)
	}

	return images.ToSlice(), nil
}

func createLogger() *zap.Logger {
	stdout := zapcore.AddSync(os.Stdout)
	verbose := viper.GetBool("verbose")

	level := zap.InfoLevel
	if verbose {
		level = zap.DebugLevel
	}

	atomicLevel := zap.NewAtomicLevelAt(level)

	loggerCfg := zap.NewProductionEncoderConfig()
	loggerCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	loggerCfg.EncodeLevel = zapcore.CapitalLevelEncoder

	return zap.New(zapcore.NewCore(
		zapcore.NewConsoleEncoder(loggerCfg),
		stdout,
		atomicLevel,
	))
}
