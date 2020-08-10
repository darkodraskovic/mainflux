// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package dbreader

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/readers/dbreader/reader"
)

var (
	errCfgNotFound     = errors.New("configuration not found")
	errFileNotFound    = errors.New("file not found")
	errWriteFile       = errors.New("failed to write file")
	errOpenFile        = errors.New("failed to open file")
	errReadFile        = errors.New("failed to read file")
	errEmptyLine       = errors.New("empty or incomplete line found in file")
	errIntervalMissing = errors.New("interval is missing")
)

// Config is used to start/stop reader and store reader data
type config struct {
	ID        string `json:"id"`
	ChannelID string `json:"channel_id"`
	scheduler *scheduler
	Reader    reader.Reader `json:"metadata"` // thing's metadata
}

// readerManager is used for handling the configurations
type readerManager struct {
	cfgFile       string
	readerType    string
	readerFactory func() reader.Reader
	configs       map[string]config
	params        map[string]string
	logger        logger.Logger
}

// NewReaderManager returns new reader manager
func NewReaderManager(cfgFile, readerType string, rf func() reader.Reader, logger logger.Logger) ReaderManager {
	return &readerManager{
		cfgFile:       cfgFile,
		readerType:    readerType,
		readerFactory: rf,
		logger:        logger,
	}
}

// ReaderManager contain the configuration to all readers
type ReaderManager interface {
	// SetParams sets reader general params used in reader init
	SetParams(map[string]string)
	// LoadAll loads reader configurations from text file
	LoadAll() error
	// SaveAll saves reader configurations to file
	SaveAll() error
	// StartAll starts all readers loaded from text file
	StartAll(Service)
	// Create creates reader configuration
	Create(string, string, string) (string, error)
	// Delete deletes reader configuration
	Delete(id string) error
	// Scheduler starts read and publish infinite loop
	Schedule(Service, string)
}

var _ ReaderManager = (*readerManager)(nil)

func (rm *readerManager) SetParams(params map[string]string) {
	rm.params = params
}

func (rm *readerManager) LoadAll() error {
	if _, err := os.Stat(rm.cfgFile); os.IsNotExist(err) {
		return errors.Wrap(errFileNotFound, err)
	}

	file, err := os.OpenFile(rm.cfgFile, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return errors.Wrap(errOpenFile, err)
	}
	defer file.Close()

	rm.configs = make(map[string]config)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		cfg := config{Reader: rm.readerFactory()}
		if err := json.Unmarshal([]byte(line), &cfg); err != nil {
			continue
		}
		cfg.Reader.Init(rm.params)

		cfg.scheduler = &scheduler{
			interval: cfg.Reader.Interval(),
			done:     make(chan bool),
		}

		rm.configs[cfg.ID] = cfg
	}
	return nil
}

func (rm *readerManager) StartAll(svc Service) {
	for id := range rm.configs {
		rm.Schedule(svc, id)
	}
}

func (rm *readerManager) SaveAll() error {
	file, err := os.Create(rm.cfgFile)
	if err != nil {
		return err
	}
	defer file.Close()

	for id := range rm.configs {
		data, err := json.Marshal(rm.configs[id])
		if err != nil {
			continue
		}

		line := fmt.Sprintf("%s\n", string(data))
		_, err = io.WriteString(file, line)
		if err != nil {
			return err
		}
	}
	return file.Sync()
}

func (rm *readerManager) Create(id string, channelID string, readerData string) (string, error) {
	reader := rm.readerFactory()
	if err := json.Unmarshal([]byte(readerData), &reader); err != nil {
		return "", err
	}
	reader.Init(rm.params)

	if cfg, ok := rm.configs[id]; ok {
		cfg.scheduler.stop()
		delete(rm.configs, id)
	}

	cfg := config{
		scheduler: &scheduler{
			interval: reader.Interval(),
			done:     make(chan bool),
		},
		ID:        id,
		ChannelID: channelID,
		Reader:    reader,
	}

	rm.configs[id] = cfg

	return id, nil
}

func (rm *readerManager) Delete(id string) error {
	cfg, ok := rm.configs[id]
	if !ok {
		return errCfgNotFound
	}

	cfg.scheduler.stop()
	delete(rm.configs, id)
	return nil
}

func (rm *readerManager) Schedule(svc Service, thingID string) {
	cfg := rm.configs[thingID]
	cfg.scheduler.start()

	go func() {
		for {
			select {
			case <-cfg.scheduler.ticker.C:
				rm.logger.Info(fmt.Sprintf("started job %s at: %s ", cfg.Reader, time.Now().Format(time.ANSIC)))
				table, err := cfg.Reader.Read()
				if err != nil {
					rm.logger.Error(err.Error())
					continue
				}
				rm.logger.Info(fmt.Sprintf("successful finished job %s at: %s ", cfg.Reader, time.Now().Format(time.ANSIC)))
				svc.Publish(
					Message{
						ThingID:   cfg.ID,
						ChannelID: cfg.ChannelID,
						Type:      rm.readerType,
						Table:     table},
				)
			case <-cfg.scheduler.done:
				return
			}
		}
	}()

	rm.logger.Info(fmt.Sprintf("Started reading %s every %f seconds", cfg.ID, cfg.scheduler.interval))
}

func validate(interval interface{}) float64 {
	if interval == nil {
		return 0
	}
	i, ok := interval.(float64)
	if !ok {
		return 0
	}

	return i
}
