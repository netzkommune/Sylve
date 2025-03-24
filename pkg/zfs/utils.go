package zfs

import (
	"bytes"
	"errors"
	"io"
	"runtime"
	"strings"
)

func (z *zfs) listByType(t, filter string) ([]*Dataset, error) {
	args := []string{"list", "-rp", "-t", t, "-o", "all"}

	if filter != "" {
		args = append(args, filter)
	}
	out, err := z.doOutput(args...)
	if err != nil {
		return nil, err
	}

	if len(out) == 0 {
		return nil, nil
	}

	var datasets []*Dataset

	name := ""
	var ds *Dataset
	for _, line := range out[1:] {
		if name != line[0] {
			name = line[0]
			ds = &Dataset{z: z, Name: name, props: make(map[string]string)}
			datasets = append(datasets, ds)
		}
		if err := ds.parseProps([][]string{out[0], line}); err != nil {
			return nil, err
		}
	}

	return datasets, nil
}

func (d *Dataset) parseProps(out [][]string) error {
	var err error

	if len(out) != 2 {
		return errors.New("output does not match what is expected on this platform")
	}
	for i, v := range out[0] {
		val := "-"
		if i < len(out[1]) {
			val = out[1][i]
		}
		d.props[strings.ToLower(v)] = val
	}

	if len(d.props) <= len(dsPropList) {
		return errors.New("output does not match what is expected on this platform")
	}
	setString(&d.Name, d.props["name"])
	setString(&d.Origin, d.props["origin"])

	if err = setUint(&d.Used, d.props["used"]); err != nil {
		return err
	}
	if err = setUint(&d.Avail, d.props["avail"]); err != nil {
		return err
	}

	setString(&d.Mountpoint, d.props["mountpoint"])
	setString(&d.Compression, d.props["compress"])
	setString(&d.Type, d.props["type"])

	if err = setUint(&d.Volsize, d.props["volsize"]); err != nil {
		return err
	}
	if err = setUint(&d.Quota, d.props["quota"]); err != nil {
		return err
	}
	if err = setUint(&d.Referenced, d.props["refer"]); err != nil {
		return err
	}

	if runtime.GOOS == "solaris" {
		return nil
	}

	if err = setUint(&d.Written, d.props["written"]); err != nil {
		return err
	}
	if err = setUint(&d.Logicalused, d.props["lused"]); err != nil {
		return err
	}
	if err = setUint(&d.Usedbydataset, d.props["usedds"]); err != nil {
		return err
	}
	return nil
}

func (z *zfs) run(in io.Reader, out io.Writer, cmd string, args ...string) ([][]string, error) {
	var stdout, stderr bytes.Buffer

	if z.sudo {
		args = append([]string{cmd}, args...)
		cmd = "sudo"
	}

	cmdOut := out
	if cmdOut == nil {
		cmdOut = &stdout
	}

	// id := uuid.New().String()
	joinedArgs := strings.Join(args, " ")

	// z.logger.Log([]string{"ID:" + id, "START", joinedArgs})
	if err := z.exec.Run(in, cmdOut, &stderr, cmd, args...); err != nil {
		return nil, &Error{
			Err:    err,
			Debug:  strings.Join([]string{cmd, joinedArgs}, " "),
			Stderr: stderr.String(),
		}
	}

	// z.logger.Log([]string{"ID:" + id, "FINISH"})

	// assume if you passed in something for stdout, that you know what to do with it
	if out != nil {
		return nil, nil
	}

	lines := strings.Split(stdout.String(), "\n")

	// last line is always blank
	lines = lines[0 : len(lines)-1]
	output := make([][]string, len(lines))

	for i, l := range lines {
		output[i] = strings.Fields(l)
	}

	return output, nil
}
