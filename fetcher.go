package main

import (
	"fmt"
	"mime"

	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"
	"github.com/emersion/go-message"
	"golang.org/x/net/html/charset"
)

type AccountInfo struct {
	FromName string
	Username string
	Password string
}

// Fetcher imap
type Fetcher interface {
	Address() string
	Search(filter Filter) (*imap.SearchData, error)
	SelectFolder(mailbox string) (*imap.SelectData, error)
	Fetch(seqSet imap.UIDSet, options *imap.FetchOptions) ([]*imapclient.FetchMessageBuffer, error)
	Exit() error
}

type fetcher struct {
	cli            *imapclient.Client
	fetcherAddress string
}

func (f *fetcher) Address() string {
	return f.fetcherAddress
}

func NewFetcher(host string, port int, conf AccountInfo) (Fetcher, error) {
	// create IMAP client
	options := &imapclient.Options{
		WordDecoder: &mime.WordDecoder{CharsetReader: charset.NewReaderLabel},
	}
	// CharsetReader
	message.CharsetReader = charset.NewReaderLabel
	c, err := imapclient.DialTLS(fmt.Sprintf("%s:%d", host, port), options)
	if err != nil {
		return nil, err
	}

	// login IMAP server
	if err = c.Login(conf.Username, conf.Password).Wait(); err != nil {
		return nil, err
	}
	return &fetcher{
		cli:            c,
		fetcherAddress: conf.Username,
	}, nil
}

func (f *fetcher) buildFetchFilter(filter Filter) *imap.SearchCriteria {
	r := &imap.SearchCriteria{
		Before: filter.Before,
		Since:  filter.Since,
	}
	if filter.MinUID > 0 || filter.MaxUID > 0 {
		uidSet := imap.UIDSet{}

		uidSet.AddRange(imap.UID(filter.MinUID), imap.UID(filter.MaxUID))
		r.UID = []imap.UIDSet{uidSet}
	}
	return r
}

func (f *fetcher) SelectFolder(mailbox string) (*imap.SelectData, error) {
	data, err := f.cli.Select(mailbox, nil).Wait()
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (f *fetcher) Search(filter Filter) (*imap.SearchData, error) {
	// search mail
	data, err := f.cli.UIDSearch(f.buildFetchFilter(filter), nil).Wait()
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (f *fetcher) Fetch(seqSet imap.UIDSet, options *imap.FetchOptions) ([]*imapclient.FetchMessageBuffer, error) {
	return f.cli.Fetch(seqSet, options).Collect()
}

func (f *fetcher) Exit() error {
	return f.cli.Logout().Wait()
}
