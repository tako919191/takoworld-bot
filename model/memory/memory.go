package memory

import (
	"bufio"
	"io"
	"strings"
)

type Memory struct {
	Name      string
	Total     string
	Usage     string
	Free      string
	Shared    string
	Buffers   string
	Available string
}

func ParseFreeCommandResultToMemory(result string) ([]Memory, error) {
	var m []Memory

	r := bufio.NewReader(strings.NewReader(result))
	for i := 0; ; i++ {
		line, err := r.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if i == 0 {
			continue
		}

		// 複数の空白を一つの空白に置き換えてからパースする。
		record := strings.Fields(line)

		var memory Memory
		memory.Name = record[0]
		memory.Total = record[1]
		memory.Usage = record[2]
		memory.Free = record[3]
		m = append(m, memory)
	}

	return m, nil
}
