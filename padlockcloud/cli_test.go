package padlockcloud

import "testing"
import "fmt"
import "io/ioutil"
import "os"
import "path/filepath"
import "time"
import "reflect"
import "gopkg.in/yaml.v2"

func NewSampleConfig(dir string) CliConfig {
	logfile := filepath.Join(dir, "LOG.txt")
	errfile := filepath.Join(dir, "ERR.txt")
	dbpath := filepath.Join(dir, "db")

	return CliConfig{
		LogConfig{
			LogFile:      logfile,
			ErrFile:      errfile,
			NotifyErrors: "notify@padlock.io",
		},
		ServerConfig{
			RequireTLS: true,
			AssetsPath: "../assets",
			Port:       5555,
			TLSCert:    "",
			TLSKey:     "",
			Host:       "myhostname.com",
		},
		LevelDBConfig{
			Path: dbpath,
		},
		EmailConfig{
			User:     "emailuser",
			Password: "emailpassword",
			Server:   "myemailserver.com",
			Port:     "4321",
		},
	}
}

func TestCliFlags(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	cfg := NewSampleConfig(dir)

	app := NewCliApp()

	go func() {
		if err := app.Run([]string{"padlock-cloud",
			"--log-file", cfg.Log.LogFile,
			"--db-path", cfg.LevelDB.Path,
			"--email-user", cfg.Email.User,
			"runserver",
			"--port", fmt.Sprintf("%d", cfg.Server.Port),
		}); err != nil {
			t.Fatal(err)
		}
	}()

	time.Sleep(time.Millisecond * 100)

	if app.Log.Config.LogFile != cfg.Log.LogFile ||
		app.Storage.Config.Path != cfg.LevelDB.Path ||
		app.Email.Config.User != cfg.Email.User ||
		app.Server.Config.Port != cfg.Server.Port {
		t.Fatal("Values provided via flags should be carried over into corresponding configs")
	}

	app.Server.Stop(time.Second)
}

func TestCliConfigFile(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	cfg := NewSampleConfig(dir)
	cfgPath := filepath.Join(dir, "config.yaml")

	yamlData, _ := yaml.Marshal(cfg)
	if err = ioutil.WriteFile(cfgPath, yamlData, 0644); err != nil {
		t.Fatal(err)
	}

	app := NewCliApp()

	go func() {
		if err := app.Run([]string{"padlock-cloud",
			"--config", cfgPath,
			"runserver",
		}); err != nil {
			t.Fatal(err)
		}
	}()

	time.Sleep(time.Millisecond * 100)

	if !reflect.DeepEqual(*app.Log.Config, cfg.Log) ||
		!reflect.DeepEqual(*app.Storage.Config, cfg.LevelDB) ||
		!reflect.DeepEqual(*app.Server.Config, cfg.Server) ||
		!reflect.DeepEqual(*app.Sender.(*EmailSender).Config, cfg.Email) {
		yamlData2, _ := yaml.Marshal(cfg)
		t.Fatalf("Config file not loaded correctly. \n\nExpected: \n\n%s\n\n Got: \n\n%s\n", yamlData, yamlData2)
	}

	app.Server.Stop(time.Second)
}
