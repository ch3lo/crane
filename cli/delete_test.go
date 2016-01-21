package cli

import (
        "testing"
	"flag"
        "github.com/stretchr/testify/assert"
	"github.com/codegangsta/cli"
)

func TestDeleteFlags(t *testing.T) {
	flags := deleteFlags()
	stringFlag, _ := flags[0].(cli.StringFlag)
	assert.Equal(t, "service-id", stringFlag.Name, "Should be \"service-id\"")
}

func TestDeleteBefore(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	set.String("service-id", "SABRE-SESSION-POOL-v1", "some hint")
	ctx := cli.NewContext(nil, set, nil)
        err := deleteBefore(ctx)
        assert.Nil(t, err, "Should be nil")
}

func TestDeleteBeforeError(t *testing.T) {
        set := flag.NewFlagSet("test", 0)
        set.String("deleteX", "SABRE-SESSION-POOL-v1", "some hint")
        ctx := cli.NewContext(nil, set, nil)
        err := deleteBefore(ctx)
        assert.NotNil(t, err, "Should throw error")
}

func TestDeleteCmd(t *testing.T) {
	stackManager = createStackManagerMock()
        set := flag.NewFlagSet("test", 0)
        set.String("deleteX", "SABRE-SESSION-POOL-v1", "some hint")
        ctx := cli.NewContext(nil, set, nil)
        deleteCmd(ctx)
}
