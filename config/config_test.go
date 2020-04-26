package config_test

import (
	"os"
	"path"
	"runtime"
	"testing"

	"github.com/spf13/afero"

	"github.com/pankrator/payment/config"
	"gopkg.in/yaml.v2"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Config Suite")
}

var _ = Describe("Config", func() {
	var c *config.Config

	_, b, _, _ := runtime.Caller(0)
	basePath := path.Dir(b)

	BeforeEach(func() {
		fs := afero.NewMemMapFs()

		file, err := fs.Create(basePath + "/config.yaml")
		Expect(err).ShouldNot(HaveOccurred())

		bytes, err := yaml.Marshal(&settings{
			Key: "file_value",
		})
		Expect(err).ShouldNot(HaveOccurred())
		_, err = file.Write(bytes)
		Expect(err).ShouldNot(HaveOccurred())
		c, err = config.New(basePath, fs)
		Expect(err).ShouldNot(HaveOccurred())
	})

	When("both env and file are set", func() {
		BeforeEach(func() {
			os.Setenv("PAY_KEY", "value")
		})
		AfterEach(func() {
			os.Unsetenv("PAY_KEY")
		})

		It("should use config from env", func() {
			s := &settings{}
			err := c.Unmarshal(s)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(s.Key).To(Equal("value"))
		})
	})

	When("env is not set", func() {
		It("should use config from file", func() {
			s := &settings{}
			err := c.Unmarshal(s)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(s.Key).To(Equal("file_value"))
		})
	})

	When("file does not exist", func() {
		It("should fail", func() {
			_, err := config.New(".", afero.NewMemMapFs())
			Expect(err).Should(HaveOccurred())
		})
	})
})

type settings struct {
	Key string `mapstructure:"key"`
}

func (s *settings) Keys() []string {
	return []string{"key"}
}
