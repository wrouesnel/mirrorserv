package entrypoint

import (
	"context"
	"fmt"
	"github.com/alecthomas/kong"
	//gap "github.com/muesli/go-app-paths"
	//"github.com/samber/lo"
	"go.uber.org/zap"
	"io"
	"mirrorserv/assets"
	"mirrorserv/pkg/kongutil"
	"mirrorserv/pkg/server"
	"mirrorserv/version"
	"os"
	"strings"
)

//nolint:gochecknoglobals
var CLI struct {
	Logging struct {
		Level  string `help:"logging level" default:"info"`
		Format string `help:"logging format (${enum})" enum:"console,json" default:"json"`
	} `embed:"" prefix:"logging."`

	Assets assets.Config `embed:"" prefix:"assets."`

	Debug struct {
		Assets struct {
			List struct {
			} `cmd:"" help:"list embedded files in the binary"`
			Cat struct {
				Filename string `arg:"" name:"filename" help:"embedded file to emit to stdout"`
			} `cmd:"" help:"output the specified file to stdout"`
		} `cmd:""`
	} `cmd:""`

	Server server.ServerConfig `cmd:"" help:"Start server"`
}

// Entrypoint is the real application entrypoint. This structure allows test packages to E2E-style tests invoking commmands
// as though they are on the command line, but using built-in coverage tools. Stub-main under the `cmd` package calls this
// function.
func Entrypoint(stdOut io.Writer, stdErr io.Writer) int {
	appCtx, appCancel := context.WithCancel(context.Background())
	defer appCancel()

	var configDirs []string
	deferredLogs := []string{}

	configfileEnvVar := fmt.Sprintf("%s_%s", strings.ToUpper(version.Name), "CONFIGFILE")
	if os.Getenv(configfileEnvVar) != "" {
		configDirs = []string{os.Getenv(configfileEnvVar)}
	} else {
		configDirs, deferredLogs = configDirListGet()
	}

	// Command line parsing can now happen
	ctx := kong.Parse(&CLI,
		kong.Description(version.Description),
		kong.Configuration(kongutil.Hybrid, configDirs...))

	// Initialize logging as soon as possible
	logConfig := zap.NewProductionConfig()
	if err := logConfig.Level.UnmarshalText([]byte(CLI.Logging.Level)); err != nil {
		deferredLogs = append(deferredLogs, err.Error())
	}
	logConfig.Encoding = CLI.Logging.Format

	logger, err := logConfig.Build()
	if err != nil {
		// Error unhandled since this is a very early failure
		_, _ = io.WriteString(stdErr, "Failure while building logger")
		return 1
	}

	// Install as the global logger
	zap.ReplaceGlobals(logger)

	// Emit deferred logs
	logger.Info("Using config paths", zap.Strings("configDirs", configDirs))
	for _, line := range deferredLogs {
		logger.Error(line)
	}

	logger.Info("Configuring asset handling", zap.Bool("use-filesystem", CLI.Assets.UseFilesystem))
	assets.UseFilesystem(CLI.Assets.UseFilesystem)

	if err := dispatchCommands(ctx, appCtx, stdOut); err != nil {
		logger.Error("Error from command", zap.Error(err))
	}

	logger.Info("Exiting normally")
	return 0
}
