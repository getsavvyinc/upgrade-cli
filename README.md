# Upgrade-CLI

<h3 align="left">
  | <a href="https://getsavvy.so/discord">Discord</a> |
</h3>


Upgrade-CLI makes it easy to add an `upgrade` command to your cli.

Upgrade-CLI was built to implement the `upgrade` command for [Savvy's](https://getsavvy.so) OSS [CLI](https://github.com/getsavvyinc/savvy-cli).

> Savvy's CLI helps developers create and share high quality runbooks right from the terminal.

## Install

```sh
go get github.com/getsavvyinc/upgrade-cli

```

## Usage

```go
package cmd

import (
	"context"
	"os"

	"github.com/getsavvyinc/savvy-cli/config"
	"github.com/getsavvyinc/savvy-cli/display"
	"github.com/getsavvyinc/upgrade-cli"
	"github.com/spf13/cobra"
)

const owner = "getsavvyinc"
const repo = "savvy-cli"

// upgradeCmd represents the upgrade command
var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "upgrade savvy to the latest version",
	Long:  `upgrade savvy to the latest version`,
	Run: func(cmd *cobra.Command, args []string) {
		executablePath, err := os.Executable()
		if err != nil {
			display.Error(err)
			os.Exit(1)
		}
		version := config.Version()

		upgrader := upgrade.NewUpgrader(owner, repo, executablePath)

		if ok, err := upgrader.IsNewVersionAvailable(context.Background(), version); err != nil {
			display.Error(err)
			return
		} else if !ok {
			display.Info("Savvy is already up to date")
			return
		}

		display.Info("Upgrading savvy...")
		if err := upgrader.Upgrade(context.Background(), version); err != nil {
			display.Error(err)
			os.Exit(1)
		} else {
			display.Success("Savvy has been upgraded to the latest version")
		}
	},
}

func init() {
	rootCmd.AddCommand(upgradeCmd)
}
```

## Requirements

> `upgrade-cli` is fully compatible with releases generated using [goreleaser](https://github.com/goreleaser/goreleaser).

`upgrade-cli` makes the following assumptions about Relase Assets.

* The checksum file has a `checksums.txt` suffix
* The checksum file format matches the example below:

```sh
6796a0fb64d0c78b2de5410a94749a3bfb77291747c1835fbd427e8bf00f6af3  savvy_darwin_arm64
3853c410eeee629f71a981844975700b2925ac7582bf5559c384c391be8abbcb  savvy_darwin_x86_64
00637eae6cf7588d990d64113a02caca831ea5391ef6f66c88db2dfa576ca6bd  savvy_linux_arm64
1e9c98dbb0f54ee06119d957fa140b42780aa330d11208ad0a21c2a06832eca3  savvy_linux_i386
3040ff4c07dda6c7ff65f9476b57277b14a72d0b33381b35aa8810df3e1785ea  savvy_linux_x86_64
```
* The URL to download a binary asset for a particular $os, $arch ends with `$os_$arch`

## Contributing

All contributions are welcome - bug reports, pull requests and ideas for improving the package.

1. Join the `#upgrade-cli` channel on [Discord](https://getsavvy.so/discord)
2. Open an [issue on GitHub](https://github.com/getsavvyinc/upgrade-cli/issues/new) to reports bugs or feature requests
3. Please follow a ["fork and pull request"](https://docs.github.com/en/get-started/exploring-projects-on-github/contributing-to-a-project) workflow for submitting changes to the repository.
