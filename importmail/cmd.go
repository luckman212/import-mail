package importmail

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
)

type Import struct {
	Options

	appendLimit int
	buffer      bytes.Buffer

	client *client.Client
}

func (c *Import) Run() (err error) {
	if len(c.Eml) == 0 {
		return errors.New("no .eml file found")
	}
	return c.run()
}
func (c *Import) run() (err error) {
	err = c.connect()
	if err != nil {
		return
	}
	defer c.disconnect()

	localLimit, err := humanize.ParseBytes(c.SizeLimit)
	if err != nil {
		return
	}
	c.appendLimit = int(localLimit)

	remoteLimit, err := c.queryAppendLimit()
	if err == nil && remoteLimit != 0 {
		c.appendLimit = humanize.IByte * int(remoteLimit)
	}
	log.Printf("APPENDLIMIT is %s", humanize.Bytes(uint64(c.appendLimit)))

	return c.doAppend()
}
func (c *Import) doAppend() error {
	for _, eml := range c.Eml {
		if c.appendLimit > 0 {
			stat, err := os.Stat(eml)
			if err != nil {
				return err
			}
			size := int(stat.Size())
			if size > c.appendLimit {
				log.Printf("skipped, %s's size %s is larger than APPENDLIMIT %s", eml, humanize.Bytes(uint64(size)), humanize.Bytes(uint64(c.appendLimit)))
				continue
			}
		}

		log.Printf("process %s", eml)

		err := c.doAppendOne(eml)
		if err != nil {
			return err
		}

		_ = os.MkdirAll(c.SaveImportedTo, 0766)
		err = os.Rename(eml, filepath.Join(c.SaveImportedTo, filepath.Base(eml)))
		if err != nil {
			return err
		}
	}

	return nil
}
func (c *Import) doAppendOne(eml string) (err error) {
	f, err := os.Open(eml)
	if err != nil {
		return
	}
	defer f.Close()
	defer c.buffer.Reset()

	scan := bufio.NewScanner(f)
	const maxBufSize = 2 * 1024 * 1024 // max line buffer 2MB
	buffer := make([]byte, maxBufSize)
	scan.Buffer(buffer, maxBufSize)
	for scan.Scan() {
		c.buffer.WriteString(scan.Text())
		c.buffer.WriteString("\r\n")
	}
	err = scan.Err()
	if err != nil {
		return
	}

	return c.client.Append(c.RemoteDir, nil, time.Time{}, &c.buffer)
}
func (c *Import) queryAppendLimit() (size uint32, err error) {
	status, err := c.client.Status(c.RemoteDir, []imap.StatusItem{imap.StatusAppendLimit})
	if err != nil {
		return
	}
	return status.AppendLimit, nil
}
func (c *Import) connect() (err error) {
	c.client, err = client.DialTLS(fmt.Sprintf("%s:%d", c.Host, c.Port), nil)
	if err == nil {
		err = c.client.Login(c.Username, c.Password)
	}
	return
}
func (c *Import) disconnect() (err error) {
	return c.client.Logout()
}
