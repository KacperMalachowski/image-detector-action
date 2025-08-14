package cmd

import (
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"github.com/KacperMalachowski/image-detector-action/pkg/strategy"
	"github.com/bmatcuk/doublestar/v4"
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


	// Exclude patterns from command line flags
	excludes := viper.GetStringSlice("exclude")

	logger.Debugw("Excluding patterns", "patterns", excludes)

	logger.Infow("Scanning directory", "directory", viper.GetString("check-directory"))

	var images []string

	filepath.WalkDir(viper.GetString("check-directory"), func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			logger.Errorw("Error walking directory", "path", path, "error", err)
			return err
		}

		// Skip directories
		if d.IsDir() {
			return nil
		}

		// Check if the file matches any exclude patterns
		for _, pattern := range excludes {
			logger.Debugw("Checking exclude pattern", "pattern", pattern, "file", path)

			matched, err := doublestar.Match(pattern, path)
			if err != nil {
				logger.Errorw("Error matching exclude pattern", "pattern", pattern, "file", path, "error", err)
				continue
			}

			if matched {
				logger.Debugw("Skipping excluded file", "pattern", pattern, "file", path)
				return nil
			}
		}

		logger.Debugw("Processing file", "file", path)

		for _, detector := range availableDetectors {
			if detector.IsSupported(path) {
				logger.Debugw("Detector found supported file", "detector", detector, "file", path)
				file, err := os.Open(path)
				if err != nil {
					logger.Errorw("Error opening file", "file", path, "error", err)
					continue
				}
				defer file.Close()

				detectedImages, err := detector.Detect(file)
				if err != nil {
					logger.Errorw("Error detecting images", "detector", detector, "file", path, "error", err)
					continue
				}

				images = append(images, detectedImages...)

				logger.Debugw("Detected images", "detector", detector, "file", path, "images", detectedImages)

				break
			}
		}

		return nil
	})

	logger.Infow("Image detection completed", "imagesFound", len(images))

	return nil
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
