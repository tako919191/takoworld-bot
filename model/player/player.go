package player

import (
	"encoding/csv"
	"io"
	"strings"
)

type Player struct {
	Name      string
	PlayerUID string
	SteamID   string
}

func ParseCSVToPlayers(csvdata string) ([]Player, error) {
	var p []Player

	r := csv.NewReader(strings.NewReader(csvdata))

	for i := 0; ; i++ {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if i == 0 {
			continue
		}

		var player Player
		player.Name = record[0]
		player.PlayerUID = record[1]
		player.SteamID = record[2]
		p = append(p, player)
	}
	return p, nil
}
