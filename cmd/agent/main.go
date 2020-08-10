package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	kubeavapis "github.com/mittwald/kube-av/pkg/apis"
	avv1beta1 "github.com/mittwald/kube-av/pkg/apis/av/v1beta1"
	"github.com/mittwald/kube-av/pkg/controller/virusscan"
	"github.com/mittwald/kube-av/pkg/engine"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedv1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

func main() {
	app := cli.App{
		Name: "kubeav-agent",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "engine",
				Usage:    "the AV engine to use",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "scan-ref",
				Usage:    "reference to the VirusScan resource",
				Required: true,
			},
			&cli.StringSliceFlag{
				Name:     "scan-dir",
				Usage:    "directories to scan",
				Required: true,
			},
		},
		Action: func(context *cli.Context) error {
			c, err := config.GetConfig()
			if err != nil {
				return err
			}

			s := scheme.Scheme

			client, err := client.New(c, client.Options{Scheme: s})
			if err != nil {
				return err
			}

			if err := kubeavapis.AddToScheme(s); err != nil {
				return err
			}

			cl, err := kubernetes.NewForConfig(c)
			if err != nil {
				return err
			}

			b := record.NewBroadcaster()
			b.StartLogging(func(format string, args ...interface{}) {
				log.Printf(format, args...)
			})
			b.StartRecordingToSink(&typedv1.EventSinkImpl{
				Interface: cl.CoreV1().Events(""),
			})

			r := b.NewRecorder(s, corev1.EventSource{Host: os.Getenv("NODE_NAME"), Component: "kubeav-agent"})

			scanRef := context.String("scan-ref")
			scanRefParts := strings.Split(scanRef, "/")
			scanDirs := context.StringSlice("scan-dir")
			engineName := context.String("engine")

			if len(scanRefParts) != 2 {
				return fmt.Errorf("invalid format for --scan-ref: %s", scanRef)
			}

			scan := avv1beta1.VirusScan{}
			scanName := types.NamespacedName{Name: scanRefParts[1], Namespace: scanRefParts[0]}

			ctx := context.Context

			if err := client.Get(ctx, scanName, &scan); err != nil {
				return err
			}

			eng, err := engine.ByName(avv1beta1.ScanEngine(engineName))
			if err != nil {
				return errors.Wrap(err, "error while loading AV engine")
			}

			result, err := eng.Execute(ctx, &scan, scanDirs)
			if err != nil {
				return err
			}

			fmt.Printf("%#v\n", scanDirs)
			fmt.Printf("%#v\n", scan)
			fmt.Printf("%+v\n", result)

			r.Eventf(&scan, corev1.EventTypeNormal, "ScanComplete", "AV scan completed")

			if len(result.InfectedFiles) > 0 {
				r.Eventf(&scan, corev1.EventTypeWarning, "InfectionFound", "found %d infected files", len(result.InfectedFiles))
			}

			patch := virusscan.PatchVirusScanResult{ScanReport: result}
			if err := client.Status().Patch(ctx, &scan, &patch); err != nil {
				return err
			}

			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		panic(err)
	}
}
