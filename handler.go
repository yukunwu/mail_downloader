package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-message/mail"
)

const (
	IMAPTimeout = 2 * time.Minute
)

type Service struct {
	Account         AccountInfo
	ImapHost        string
	ImapPort        int
	OutputDirectory string
}

func (s *Service) AutoFetch(ctx context.Context, folder string, startTime, endTime string, minUID, maxUID uint32, batchSize int) error {
	filter := Filter{
		Folder: folder,
	}
	if minUID > 0 {
		filter.MinUID = minUID
	}
	if maxUID > 0 {
		filter.MaxUID = maxUID
	}
	if len(startTime) > 0 {
		t, err := time.Parse(TimeLayoutStr, startTime)
		if err != nil {
			return err
		}
		filter.Since = t
	}

	if len(endTime) > 0 {
		t, err := time.Parse(TimeLayoutStr, endTime)
		if err != nil {
			return err
		}
		filter.Before = t
	}
	if err := s.fetchWithFolder(ctx, filter, batchSize); err != nil {
		return err
	}
	return nil
}

func (s *Service) fetchWithFolder(ctx context.Context, filter Filter, batchSize int) error {
	log.Printf("%s Start fetching email,mailbox:%s\n", time.Now().Format(TimeLayoutStr), filter.Folder)
	err := s.FetchAndSave(ctx, batchSize, filter)
	if err != nil {
		return err
	}
	log.Printf("%s Finished fetching email,mailbox:%s\n", time.Now().Format(TimeLayoutStr), filter.Folder)
	return nil
}
func (s *Service) FetchAndSave(ctx context.Context, batchSize int, filter Filter) error {
	var (
		emailFolder = filepath.Join(s.OutputDirectory, s.Account.Username, filter.Folder)
	)
	if err := os.MkdirAll(emailFolder, os.ModePerm); err != nil {
		return err
	}
	emailClient, err := NewFetcher(s.ImapHost, s.ImapPort, s.Account)
	if err != nil {
		return err
	}
	defer emailClient.Exit()
	clientStartTime := time.Now()
	_, err = emailClient.SelectFolder(filter.Folder)
	if err != nil {
		return nil
	}
	// Search Data
	data, err := emailClient.Search(filter)
	if err != nil {
		return err
	}
	if data.All == nil {
		return nil
	}
	allUIDArr := chunkBy(data.AllUIDs(), batchSize)
	for _, ids := range allUIDArr {
	ImapReLogin:
		if time.Since(clientStartTime) > IMAPTimeout {
			emailClient, err = NewFetcher(s.ImapHost, s.ImapPort, s.Account)
			if err != nil {
				return err
			}
			clientStartTime = time.Now()
			_, err = emailClient.SelectFolder(filter.Folder)
			if err != nil {
				return err
			}
		}

		fetchOptions := &imap.FetchOptions{
			UID:         true,
			BodySection: []*imap.FetchItemBodySection{{}},
		}
		messages, err1 := emailClient.Fetch(imap.UIDSetNum(ids...), fetchOptions)
		if err1 != nil {
			if strings.Contains(err1.Error(), "use of closed network connection") {
				log.Printf("emailClient.Fetch:%s\n", err1.Error())
				clientStartTime = time.Now().Add(-IMAPTimeout)
				goto ImapReLogin
			}
			log.Printf("emailClient.Fetch,error:%v\n", err1)
			continue
		}
		for _, msg := range messages {
			t := &EmailInfo{
				To:     s.Account.Username,
				Folder: filter.Folder,
				UID:    uint32(msg.UID),
			}
			for _, v := range msg.BodySection {
				if err = s.parseBodySection(bytes.NewBuffer(v), t); err != nil {
					log.Printf("parseBodySection,error:%v\n", err)
					continue
				}
				dateFolder := filepath.Join(emailFolder, t.Date.Format("2006-01-02"))
				if err = os.MkdirAll(dateFolder, os.ModePerm); err != nil {
					log.Printf(" os.MkdirAll,error:%v\n", err)
					continue
				}
				fileName := filepath.Join(dateFolder, strings.ReplaceAll(t.Subject, "/", "_")+".eml")
				fileIdx := 0
				for {
					if _, err2 := os.Stat(fileName); errors.Is(err2, os.ErrNotExist) {
						break
					}
					fileIdx++
					fileName = filepath.Join(dateFolder, fmt.Sprintf("%s_%d.eml", strings.ReplaceAll(t.Subject, "/", "_"), fileIdx))
				}
				f, errF := os.Create(fileName)
				if errF != nil {
					log.Printf("os.Create,error:%v\n", errF)
					continue
				}
				if _, errF = f.Write(v); errF != nil {
					log.Printf("Write File,error:%v\n", errF)
					_ = f.Close()
					continue
				}
				_ = f.Close()
				log.Printf("%s finish:%s\n", time.Now().Format(TimeLayoutStr), f.Name())
			}
		}
	}
	return nil
}

func (s *Service) parseBodySection(reader io.Reader, t *EmailInfo) error {
	mr, err := mail.CreateReader(reader)
	if err != nil {
		return err
	}
	defer mr.Close()

	header := mr.Header
	if date, err := header.Date(); err == nil {
		t.Date = date
	}
	if from, err := header.AddressList("From"); err == nil && len(from) > 0 {
		t.From = strings.Join(extract(len(from), func(i int) string {
			return from[i].Address
		}), ",")
	}
	if subject, err := header.Subject(); err == nil {
		t.Subject = subject
	}
	return nil
}
