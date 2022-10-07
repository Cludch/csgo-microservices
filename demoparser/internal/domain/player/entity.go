package player

import (
	"time"

	"github.com/Cludch/csgo-microservices/shared/pkg/entity"
	"github.com/go-playground/validator"
)

var validate = validator.New()

type Player struct {
	ID        uint64          `json:"id" bson:"_id" validate:"required"`
	CreatedAt time.Time       `json:"-" bson:"createdAt"`
	FaceitID  entity.ID       `json:"faceitId" bson:"faceitId,omitempty"`
	Results   []*PlayerResult `json:"results" bson:"results" validation:"dive"`
}

// PlayerResult holds different performance metrics from one game.
type PlayerResult struct {
	MatchID             entity.ID `json:"matchId" bson:"matchId" validation:"required"`
	Time                time.Time `json:"time" bson:"time" validate:"required"`
	Map                 string    `json:"map" bson:"map" validate:"required"`
	MatchRounds         byte      `json:"matchRounds" bson:"matchRounds" validation:"required"`
	ScoreOwnTeam        byte      `json:"scoreOwnTeam" bson:"scoreOwnTeam"`
	ScoreEnemyTeam      byte      `json:"scoreEnemyTeam" bson:"scoreEnemyTeam"`
	SteamID             uint64    `json:"id" bson:"steamId" validation:"required"`
	Name                string    `json:"name" bson:"name" validation:"required"`
	Kills               int8      `json:"kills" bson:"kills"`
	EntryKills          byte      `json:"entryKills" bson:"entryKills"`
	OpeningDuelAttempts byte      `json:"openingDuelAttempts" bson:"openingDuelAttempts"`
	Headshots           byte      `json:"headshots" bson:"headshots"`
	Assists             byte      `json:"assists" bson:"assists"`
	Deaths              byte      `json:"deaths" bson:"deaths"`
	DamageDealt         int       `json:"damageDealt" bson:"damageDealt"`
	MVPs                byte      `json:"mvps" bson:"mvps"`
	Won1v3              byte      `json:"won1v3" bson:"won1v3"`
	Won1v4              byte      `json:"won1v4" bson:"won1v4"`
	Won1v5              byte      `json:"won1v5" bson:"won1v5"`
	RoundsWith3K        byte      `json:"roundsWith3k" bson:"3k"`
	RoundsWith4K        byte      `json:"roundsWith4k" bson:"4k"`
	RoundsWith5K        byte      `json:"roundsWith5k" bson:"5k"`
	// Rank
	WinCount int `json:"wins" bson:"winCount"`
	RankOld  int `json:"rankOld" bson:"rankOld"`
	RankNew  int `json:"rankNew" bson:"rankNew"`
}

func NewPlayer(id uint64) (*Player, error) {
	p := &Player{
		ID:        id,
		CreatedAt: time.Now(),
		Results:   []*PlayerResult{},
	}

	if err := p.Validate(); err != nil {
		return nil, err
	}
	return p, nil
}

func (p *Player) Validate() error {
	err := validate.Struct(p)
	if err != nil {
		return err.(validator.ValidationErrors)
	}

	return nil
}
