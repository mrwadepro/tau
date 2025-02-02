package compile

import (
	"errors"
	"fmt"
	"io"

	"github.com/pterm/pterm"
	spec "github.com/taubyte/go-specs/common"
	"github.com/taubyte/go-specs/methods"
	"github.com/taubyte/tau/protocols/monkey/jobs"
)

func (ctx resourceContext) stashAndPush(id string, file io.ReadSeekCloser) error {
	if file == nil {
		return errors.New("file is nil")
	}

	tnsClient, err := ctx.simple.TNS()
	if err != nil {
		return err
	}

	c := jobs.Context{
		Tns:  tnsClient,
		Node: ctx.universe.TNS().Node(),
	}
	c.ForceContext(ctx.universe.Context())

	cid, err := c.StashBuildFile(file)
	if err != nil {
		return fmt.Errorf("stash failed with: %s", err)
	}

	assetKey, err := methods.GetTNSAssetPath(ctx.projectId, id, spec.DefaultBranch)
	if err != nil {
		return err
	}

	pterm.Info.Printf("Stashing file to: %s => %s\n", assetKey.String(), cid)

	err = c.Tns.Push(assetKey.Slice(), cid)
	if err != nil {
		return fmt.Errorf("saving asset file failed with")
	}

	return nil
}
