package convert

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/amit7itz/goset"
	"github.com/otterize/intents-operator/src/operator/api/v1alpha2"
	"github.com/otterize/otterize-cli/src/pkg/consts"
	"github.com/otterize/otterize-cli/src/pkg/intentsoutput"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8syaml "k8s.io/apimachinery/pkg/util/yaml"
	"os"
	"path/filepath"
	"sigs.k8s.io/yaml"
)

const regularFile = 0
const FilepathKey = "filename"
const FilepathShorthand = "f"

func NewIntentsResourceFromIntentsSpec(spec v1alpha2.IntentsSpec) *v1alpha2.ClientIntents {
	return &v1alpha2.ClientIntents{
		TypeMeta: v1.TypeMeta{
			Kind:       consts.IntentsKind,
			APIVersion: consts.IntentsAPIVersion,
		},
		ObjectMeta: v1.ObjectMeta{
			Name: spec.Service.Name,
		},
		Spec: spec.DeepCopy(),
	}
}

var ConvertCmd = &cobra.Command{
	Use:   "convert",
	Short: "Converts Otterize intents to Kubernetes ClientIntents resources",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		printer := intentsoutput.IntentsPrinter{}
		allowedExts := goset.NewSet(".yaml", ".yml")
		fileInfo, err := os.Stat(viper.GetString(FilepathKey))
		if err != nil {
			return fmt.Errorf("failed to get info for path %s: %w", viper.GetString(FilepathKey), err)
		}
		filePaths := make([]string, 0)
		if fileInfo.IsDir() {
			entries, err := os.ReadDir(viper.GetString(FilepathKey))
			if err != nil {
				return fmt.Errorf("failed to read dir %s: %w", viper.GetString(FilepathKey), err)
			}
			for _, entry := range entries {
				if !allowedExts.Contains(filepath.Ext(entry.Name())) || entry.Type() != regularFile {
					continue
				}
				filePaths = append(filePaths, filepath.Join(viper.GetString(FilepathKey), entry.Name()))
			}
		} else {
			filePaths = append(filePaths, viper.GetString(FilepathKey))
		}

		for _, path := range filePaths {
			err := func() error {
				file, err := os.Open(path)
				if err != nil {
					return err
				}
				defer file.Close()
				yamlReader := k8syaml.NewYAMLReader(bufio.NewReader(file))
				for {
					doc, err := yamlReader.Read()
					if err != nil {
						if errors.Is(err, io.EOF) {
							break
						}

						return fmt.Errorf("unable to parse YAML file %s: %w", path, err)
					}

					var intentsSpec v1alpha2.IntentsSpec
					err = yaml.UnmarshalStrict(doc, &intentsSpec)
					if err != nil {
						return fmt.Errorf("unable to parse YAML file %s: %w", path, err)
					}

					resource := NewIntentsResourceFromIntentsSpec(intentsSpec)
					err = printer.PrintObj(resource, os.Stdout)
					if err != nil {
						return err
					}
				}
				return nil
			}()
			if err != nil {
				return err
			}
		}

		return nil
	},
}

func init() {
	ConvertCmd.Flags().StringP(FilepathKey, FilepathShorthand, ".",
		"filename that contains the intents, or a directory containing intents")
	cobra.CheckErr(ConvertCmd.MarkFlagRequired(FilepathKey))
}
