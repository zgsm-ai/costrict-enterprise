package env

import "testing"

func TestEnv(t *testing.T) {
	var test bool
	envs := NewEnvs()
	envs.Register("smc_TEST", "test", "test", "ENABLE", NewBool(&test))
	envs.SetOnChange(func() error {
		return envs.Save("testenv.env")
	})
	envs.Load("testenv.env")
	if !test {
		t.FailNow()
	}
	if err := envs.SetEnv("smc_TEST", "DISABLE"); err != nil {
		t.FailNow()
	}
	if test {
		t.FailNow()
	}
	if err := envs.SetEnv("smc_TEST", "ENABLE"); err != nil {
		t.FailNow()
	}
	if !test {
		t.FailNow()
	}
}
