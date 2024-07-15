package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/wantedly/container-platform-tools/dockerplatforms"
	"github.com/wantedly/container-platform-tools/k8splatforms"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func main() {
	var c = cmdargs{}
	var rootCmd = &cobra.Command{
		Use: "kubectl-platforms image...",
		RunE: func(cmd *cobra.Command, args []string) error {
			c.stdout = cmd.OutOrStdout()
			c.stderr = cmd.ErrOrStderr()
			return c.Run(cmd.Context())
		},
	}

	var kubeconfigDefault string
	if home := homedir.HomeDir(); home != "" {
		kubeconfigDefault = filepath.Join(home, ".kube", "config")
	}
	rootCmd.PersistentFlags().StringVar(&c.kubeconfig, "kubeconfig", kubeconfigDefault, "Path to the kubeconfig file")
	rootCmd.PersistentFlags().StringVar(&c.after, "after", "", "Take into account resources after this time (RFC3339)")
	c.nodePlatforms = dockerPlatformList(dockerplatforms.MustParseDockerPlatformList("linux/amd64"))
	rootCmd.PersistentFlags().Var(&c.nodePlatforms, "node-platforms", "List of node platforms")
	rootCmd.PersistentFlags().BoolVar(&c.csv, "csv", false, "Output in CSV format")

	err := rootCmd.Execute()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

type cmdargs struct {
	stdout        io.Writer
	stderr        io.Writer
	kubeconfig    string
	after         string
	nodePlatforms dockerPlatformList
	csv           bool
}

type dockerPlatformList dockerplatforms.DockerPlatformList

func (l *dockerPlatformList) String() string {
	return dockerplatforms.DockerPlatformList(*l).String()
}
func (l *dockerPlatformList) Set(text string) error {
	list, err := dockerplatforms.ParseDockerPlatformList(text)
	if err != nil {
		return errors.Wrap(err, "parsing docker platform list")
	}
	*l = dockerPlatformList(list)
	return nil
}
func (l *dockerPlatformList) Type() string {
	return "dockerPlatformList"
}

func (c *cmdargs) Run(ctx context.Context) error {
	resolver := dockerplatforms.NewImageTools()
	cache, err := dockerplatforms.NewYAMLCache(ctx, "image-platforms.yaml")
	if err != nil {
		return errors.Wrap(err, "initializing cache")
	}
	defer cache.WriteBack(ctx)
	inspector := dockerplatforms.New(resolver, cache)

	config, err := clientcmd.BuildConfigFromFlags("", c.kubeconfig)
	if err != nil {
		return errors.Wrap(err, "loading kubeconfig")
	}

	after, err := time.Parse(time.RFC3339, c.after)
	if err != nil {
		return errors.Wrap(err, "parsing time")
	}

	rows, collectErr := k8splatforms.Collector{
		RESTConfig:        config,
		After:             after,
		NodePlatforms:     dockerplatforms.DockerPlatformList(c.nodePlatforms),
		PlatformInspector: inspector,
		Processors: []k8splatforms.KindProcessor{
			k8splatforms.PodProcessor{},
			k8splatforms.ReplicaSetProcessor{},
			k8splatforms.DeploymentProcessor{},
			k8splatforms.StatefulSetProcessor{},
			k8splatforms.DaemonSetProcessor{},
			k8splatforms.JobProcessor{},
			k8splatforms.CronJobProcessor{},
			k8splatforms.WorkflowProcessor{},
			k8splatforms.CronWorkflowProcessor{},
		},
	}.Collect(ctx)

	if collectErr != nil && len(rows) == 0 {
		return errors.Wrap(collectErr, "collecting platforms")
	}

	if c.csv {
		writer := csv.NewWriter(c.stdout)
		err := writer.Write([]string{
			"Namespace",
			"APIVersion",
			"Kind",
			"Name",
			"SubName",
			"ScheduledPlatform",
			"DeclaredPlatforms",
			"ImagePlatforms",
			"ImagePlatformDetails",
			"HasViolation",
			"CPUUsage",
			"MemoryUsage",
			"Error",
		})
		if err != nil {
			return errors.Wrap(err, "writing CSV header")
		}
		for _, row := range rows {
			var scheduledPlatform string
			if row.ScheduledPlatform != nil {
				scheduledPlatform = row.ScheduledPlatform.String()
			}
			imagePlatformDetailsMap := make(map[string]string)
			for k, v := range row.ImagePlatformDetails {
				imagePlatformDetailsMap[k] = v.String()
			}
			imagePlatformDetails, err := json.Marshal(imagePlatformDetailsMap)
			if err != nil {
				return errors.Wrap(err, "marshaling image platform details")
			}
			err = writer.Write([]string{
				row.Namespace,
				row.APIVersion,
				row.Kind,
				row.Name,
				row.SubName,
				scheduledPlatform,
				row.DeclaredPlatforms.String(),
				row.ImagePlatforms.String(),
				string(imagePlatformDetails),
				fmt.Sprintf("%v", row.HasViolation),
				fmt.Sprintf("%v", row.CPUUsage),
				fmt.Sprintf("%v", row.MemoryUsage),
				row.Error,
			})
			if err != nil {
				return errors.Wrap(err, "writing CSV row")
			}
		}

		writer.Flush()
		err = writer.Error()
		if err != nil {
			return errors.Wrap(err, "flushing CSV")
		}
	} else {
		stats := make(map[string]counts)
		allKey := "All"
		stats[allKey] = counts{}
		for _, platform := range c.nodePlatforms {
			stats[allKey] = counts{}
			stats[fmt.Sprintf("DeclaredPlatform including %s", platform)] = counts{}
			stats[fmt.Sprintf("ImagePlatform including %s", platform)] = counts{}
			stats[fmt.Sprintf("ScheduledPlatform = %s", platform)] = counts{}
			stats["ScheduledPlatform = (pending)"] = counts{}
			stats["HasViolation"] = counts{}
		}
		for _, row := range rows {
			isPod := row.APIVersion == "v1" && row.Kind == "Pod"
			rowCount := counts{}
			if isPod {
				rowCount.numPods++
				rowCount.cpuSum += row.CPUUsage
				rowCount.memorySum += row.MemoryUsage
			} else {
				rowCount.numNonPods++
			}
			stats[allKey] = stats[allKey].Add(rowCount)

			declKey := fmt.Sprintf("DeclaredPlatforms = %s", row.DeclaredPlatforms)
			if len(row.DeclaredPlatforms) == 0 {
				declKey = "DeclaredPlatforms = (empty)"
			}
			stats[declKey] = stats[declKey].Add(rowCount)
			for _, platform := range row.DeclaredPlatforms {
				key := fmt.Sprintf("DeclaredPlatform including %s", platform)
				stats[key] = stats[key].Add(rowCount)
			}

			imgKey := fmt.Sprintf("ImagePlatforms = %s", row.DeclaredPlatforms)
			if len(row.DeclaredPlatforms) == 0 {
				imgKey = "ImagePlatforms = (empty)"
			}
			stats[imgKey] = stats[imgKey].Add(rowCount)
			for _, platform := range row.ImagePlatforms {
				key := fmt.Sprintf("ImagePlatform including %s", platform)
				stats[key] = stats[key].Add(rowCount)
			}

			if row.ScheduledPlatform != nil {
				key := fmt.Sprintf("ScheduledPlatform = %s", row.ScheduledPlatform)
				stats[key] = stats[key].Add(rowCount)
			} else if isPod {
				key := "ScheduledPlatform = (pending)"
				stats[key] = stats[key].Add(rowCount)
			}

			if row.HasViolation {
				key := "HasViolation"
				stats[key] = stats[key].Add(rowCount)
			}
		}

		keys := make([]string, 0, len(stats))
		for key := range stats {
			keys = append(keys, key)
		}
		slices.Sort(keys)

		allCounts := stats[allKey]
		for _, key := range keys {
			_, err := fmt.Fprintf(c.stdout, "%s:\n", key)
			if err != nil {
				return errors.Wrap(err, "writing stats")
			}
			if allCounts.numPods > 0 {
				_, err := fmt.Fprintf(c.stdout, "  Pods: %d / %d (%v%%)\n", stats[key].numPods, allCounts.numPods, 100*float64(stats[key].numPods)/float64(allCounts.numPods))
				if err != nil {
					return errors.Wrap(err, "writing stats")
				}
			}
			if allCounts.numNonPods > 0 {
				_, err := fmt.Fprintf(c.stdout, "  Non-pods: %d / %d (%v%%)\n", stats[key].numNonPods, allCounts.numNonPods, 100*float64(stats[key].numNonPods)/float64(allCounts.numNonPods))
				if err != nil {
					return errors.Wrap(err, "writing stats")
				}
			}
			if allCounts.cpuSum > 0 {
				_, err := fmt.Fprintf(c.stdout, "  CPU: %v / %v (%v%%)\n", stats[key].cpuSum, allCounts.cpuSum, 100*stats[key].cpuSum/allCounts.cpuSum)
				if err != nil {
					return errors.Wrap(err, "writing stats")
				}
			}
			if allCounts.memorySum > 0 {
				_, err := fmt.Fprintf(c.stdout, "  Memory: %v / %v (%v%%)\n", stats[key].memorySum, allCounts.memorySum, 100*stats[key].memorySum/allCounts.memorySum)
				if err != nil {
					return errors.Wrap(err, "writing stats")
				}
			}
		}

		fmt.Fprintf(c.stdout, "Violations:\n")
		for _, row := range rows {
			if row.HasViolation {
				if row.SubName != "" {
					_, err := fmt.Fprintf(c.stdout, "%s:%s.%s/%s(%s):\n", row.Namespace, row.APIVersion, row.Kind, row.Name, row.SubName)
					if err != nil {
						return errors.Wrap(err, "writing stats")
					}
				} else {
					_, err := fmt.Fprintf(c.stdout, "%s:%s.%s/%s:\n", row.Namespace, row.APIVersion, row.Kind, row.Name)
					if err != nil {
						return errors.Wrap(err, "writing stats")
					}
				}
				containerKeys := make([]string, 0, len(row.ImagePlatformDetails))
				for key := range row.ImagePlatformDetails {
					containerKeys = append(containerKeys, key)
				}
				slices.Sort(containerKeys)
				for _, containerKey := range containerKeys {
					platforms := row.ImagePlatformDetails[containerKey]
					if !platforms.Includes(row.DeclaredPlatforms) {
						_, err := fmt.Fprintf(c.stdout, "  %s: %s\n", containerKey, platforms)
						if err != nil {
							return errors.Wrap(err, "writing stats")
						}
					}
				}
			}
		}
	}

	return errors.Wrap(collectErr, "collecting platforms")
}

type counts struct {
	numPods    int
	numNonPods int
	cpuSum     float64
	memorySum  float64
}

func (c counts) Add(other counts) counts {
	c.numPods += other.numPods
	c.numNonPods += other.numNonPods
	c.cpuSum += other.cpuSum
	c.memorySum += other.memorySum
	return c
}
