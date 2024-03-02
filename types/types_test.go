package types

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestParsesEmulation(t *testing.T) {
	path := filepath.Join("..", "test", "testdata", "emulation.yml")

	// Read the YAML file
	yamlFile, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read %s: %s", path, err)
	}

	// Unmarshal the YAML file into an Emulation struct
	var emulation Emulation
	err = yaml.Unmarshal(yamlFile, &emulation)
	if err != nil {
		t.Fatalf("Unmarshal failed: %s", err)
	}

	assert.Equal(t, 2, len(emulation.Atomics))

	t1048 := emulation.Atomics[0]
	assert.Equal(t, 1, len(t1048.AtomicTests))

	var signatures []string = *t1048.AtomicTests[0].Signatures
	assert.Equal(t, 1, len(signatures))

	t1543 := emulation.Atomics[1]
	assert.Equal(t, 1, len(t1543.AtomicTests))

	signatures = *t1543.AtomicTests[0].Signatures
	assert.Equal(t, 2, len(signatures))

	sigsOnly := *emulation.SignaturesOnly
	assert.Equal(t, true, sigsOnly)
}
