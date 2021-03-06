package main

import (
	"fmt"
	"os"

	"github.com/rancher/log"
	"github.com/rancher/vxlan/server"
	"github.com/rancher/vxlan/vxlan"
	"github.com/urfave/cli"
)

var (
	// VERSION Of the binary
	VERSION = "0.0.0-dev"
)

func main() {
	app := cli.NewApp()
	app.Version = VERSION
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "listen",
			Value:  "127.0.0.1:8111",
			EnvVar: "RANCHER_SERVICE_LISTEN_PORT",
		},
		cli.BoolFlag{
			Name:   "debug, d",
			EnvVar: "RANCHER_DEBUG",
		},
		cli.StringFlag{
			Name:   "bridge",
			Value:  vxlan.DefaultBridgeName,
			EnvVar: "RANCHER_BRIDGE",
		},
		cli.StringFlag{
			Name:   "metadata-address",
			Value:  vxlan.DefaultMetadataAddress,
			EnvVar: "RANCHER_METADATA_ADDRESS",
		},
		cli.IntFlag{
			Name:   "vtep-mtu",
			Value:  vxlan.DefaultVxlanMTU,
			EnvVar: "RANCHER_VTEP_MTU",
		},
		cli.IntFlag{
			Name:   "vxlan-vni",
			Value:  vxlan.DefaultVxlanVNI,
			EnvVar: "RANCHER_VXLAN_VNI",
		},
	}
	app.Action = func(ctx *cli.Context) {
		if err := appMain(ctx); err != nil {
			log.Errorf("error: %v", err)
			os.Exit(1)
		}
	}

	app.Run(os.Args)
}

func appMain(ctx *cli.Context) error {
	if ctx.Bool("debug") {
		log.SetLevelString("debug")
	}

	done := make(chan error)

	v, err := vxlan.NewVxlan(ctx.String("metadata-address"))
	if err != nil {
		return err
	}
	v.BridgeName = ctx.String("bridge")
	v.VxlanMTU = ctx.Int("vtep-mtu")
	v.VxlanVNI = ctx.Int("vxlan-vni")
	v.VxlanInterfaceName = fmt.Sprintf("vtep%d", v.VxlanVNI)
	err = v.SetDefaultVxlanInterfaceInfo()
	if err != nil {
		return err
	}
	err = v.Start()
	if err != nil {
		return err
	}

	listenPort := ctx.String("listen")
	log.Debugf("About to start server and listen on port: %v", listenPort)
	go func() {
		s := server.Server{V: v}
		done <- s.ListenAndServe(listenPort)
	}()

	return <-done
}
