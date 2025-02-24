package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"strings"

	dem "github.com/markus-wa/demoinfocs-golang/v3/pkg/demoinfocs"
	common "github.com/markus-wa/demoinfocs-golang/v3/pkg/demoinfocs/common"
	events "github.com/markus-wa/demoinfocs-golang/v3/pkg/demoinfocs/events"
)

const unknown = "Unknown"
const spectator = "Spectator"
const unassigned = "Unassigned"
const plant = "plant"
const unranked = "Unranked"
const zeroTime = "00:00"



// Game is the overall struct that holds the parsed demo data.
type Game struct {
	MatchName      string          `json:"matchID"`
	ClientName     string          `json:"clientName"`
	Map            string          `json:"mapName"`
	TickRate       int64           `json:"tickRate"`
	PlaybackTicks  int64           `json:"playbackTicks"`
	PlaybackFrames int64           `json:"playbackFramesCount"`
	ParsedToFrame  int64           `json:"parsedToFrameIdx"`
	ParsingOpts    ParserOpts      `json:"parserParameters"`
	ServerVars     ServerConVar    `json:"serverVars"`
	MatchPhases    MatchPhases     `json:"matchPhases"`
	Connections    []ConnectAction `json:"playerConnections"`
	Rounds         []GameRound     `json:"gameRounds"`
}

// ParserOpts holds the parameters passed to the parser.
type ParserOpts struct {
	ParseRate       int    `json:"parseRate"`
	ParseFrames     bool   `json:"parseFrames"`
	ParseKillFrames bool   `json:"parseKillFrames"`
	TradeTime       int64  `json:"tradeTime"`
	RoundBuyStyle   string `json:"roundBuyStyle"`
	DamagesRolled   bool   `json:"damagesRolledUp"`
}

// MatchPhases holds lists of when match events occurred.
type MatchPhases struct {
	AnnLastRoundHalf    []int64 `json:"announcementLastRoundHalf"`
	AnnFinalRound       []int64 `json:"announcementFinalRound"`
	AnnMatchStarted     []int64 `json:"announcementMatchStarted"`
	RoundStarted        []int64 `json:"roundStarted"`
	RoundEnded          []int64 `json:"roundEnded"`
	RoundFreezeEnded    []int64 `json:"roundFreezetimeEnded"`
	//potentially at end of buytime
	RoundEndedOfficial  []int64 `json:"roundEndedOfficial"`
	GameHalfEnded       []int64 `json:"gameHalfEnded"`
	MatchStart          []int64 `json:"matchStart"`
	MatchStartedChanged []int64 `json:"matchStartedChanged"`
	WarmupChanged       []int64 `json:"warmupChanged"`
	TeamSwitch          []int64 `json:"teamSwitch"`
}

// ServerConVar holds server convars, like round timers and timeouts, etc. Not always accurate.
//todo: where are the values if not always accurate? 
type ServerConVar struct {
	CashBombDefused               int64 `json:"cashBombDefused"`               // cash_player_bomb_defused
	CashBombPlanted               int64 `json:"cashBombPlanted"`               // cash_player_bomb_planted
	CashWinBomb                   int64 `json:"cashTeamTWinBomb"`              // cash_team_terrorist_win_bomb
	CashWinDefuse                 int64 `json:"cashWinDefuse"`                 // cash_team_win_by_defusing_bomb
	CashWinTimeRunOut             int64 `json:"cashWinTimeRunOut"`             // cash_team_win_by_time_running_out_bomb
	CashWinElimination            int64 `json:"cashWinElimination"`            // cash_team_elimination_bomb_map
	CashPlayerKilledDefault       int64 `json:"cashPlayerKilledDefault"`       // cash_player_killed_enemy_default
	CashTeamLoserBonus            int64 `json:"cashTeamLoserBonus"`            // cash_team_loser_bonus
	CashTeamLoserBonusConsecutive int64 `json:"cashTeamLoserBonusConsecutive"` // cash_team_loser_bonus_consecutive_rounds
	RoundTime                     int64 `json:"roundTime"`                     // mp_roundtime
	RoundTimeDefuse               int64 `json:"roundTimeDefuse"`               // mp_roundtime_defuse
	RoundRestartDelay             int64 `json:"roundRestartDelay"`             // mp_round_restart_delay
	FreezeTime                    int64 `json:"freezeTime"`                    // mp_freezetime
	BuyTime                       int64 `json:"buyTime"`                       // mp_buytime
	BombTimer                     int64 `json:"bombTimer"`                     // mp_c4timer
	MaxRounds                     int64 `json:"maxRounds"`                     // mp_maxrounds
	TimeoutsAllowed               int64 `json:"timeoutsAllowed"`               // mp_team_timeout_max
	CoachingAllowed               int64 `json:"coachingAllowed"`               // sv_coaching_enabled
}





// ConnectAction is the act of connecting or disconnecting to the server.
type ConnectAction struct {
	Tick        int64  `json:"tick"`
	ConnectType string `json:"action"`
	SteamID     uint64 `json:"steamID"`
}

// GameRound contains round info and events.
type GameRound struct {
	RoundNum             int64              `json:"roundNum"`
	IsWarmup             bool               `json:"isWarmup"`
	StartTick            int64              `json:"startTick"`
	FreezeTimeEndTick    int64              `json:"freezeTimeEndTick"`
	EndTick              int64              `json:"endTick"`
	EndOfficialTick      int64              `json:"endOfficialTick"`
	TScore               int64              `json:"tScore"`
	CTScore              int64              `json:"ctScore"`
	EndTScore            int64              `json:"endTScore"`
	EndCTScore           int64              `json:"endCTScore"`
	CTTeam               *string            `json:"ctTeam"`
	TTeam                *string            `json:"tTeam"`
	WinningSide          string             `json:"winningSide"`
	WinningTeam          *string            `json:"winningTeam"`
	LosingTeam           *string            `json:"losingTeam"`
	Reason               string             `json:"roundEndReason"`
	CTSide               PlayerTeams        `json:"ctSide"`
	TSide                PlayerTeams        `json:"tSide"`
	Economy		         []PlayerEconomy	`json:"Economy"`
	Frames               []GameFrame        `json:"frames"`
}

// PlayerTeam.
type PlayerTeams struct {
	TeamName string    `json:"teamName"`
	Players  []Players `json:"players"`
}

// Players.
type Players struct {
	PlayerName string `json:"playerName"`
	SteamID    int64  `json:"steamID"`
}


type PlayerEconomy struct {
	Name 		string		`json:"Name"`
	SteamID   	int64       `json:"steamID"`
	Team      	String      `json:"team"`
	Side	    string      `json:"side"`
	EqValTemp   int64       `json:"equipmentValueTemp"`
	EqValFreeze int64       `json:"equipmentValueFreezetimeEnd"`
	EqValStart  int64       `json:"equipmentValueRoundStart"`
	EqVal		int64       `json:"equipmentValueRoundEnd"`
	Moneytemp	int64       `json:"cashTemp"`
	MoneyEarned	int64       `json:"cashEarned"`
	MoneyStart	int64       `json:"cashStart"`
	MoneyFreeze	int64       `json:"cashFreezetimeEnd"`
	Money		int64       `json:"cashEnd"`
	MoneySpent	int64       `json:"cashSpent"`
}




// GameFrame (game state at time t).
type GameFrame struct {
	FrameID       int64         `json:"frameID"`
	GlobalFrameID int64         `json:"globalFrameID"`
	IsKillFrame   bool          `json:"isKillFrame"`
	Tick          int64         `json:"tick"`
	T             TeamFrameInfo `json:"t"`
	CT            TeamFrameInfo `json:"ct"`
}


// TeamFrameInfo at time t.
type TeamFrameInfo struct {
	Side         string       `json:"side"`
	Team         string       `json:"teamName"`
	CurrentEqVal int64        `json:"teamEqVal"`
	AlivePlayers int64        `json:"alivePlayers"`
	RoundReward  int64        `json:"RoundReward"`
	Players      []PlayerInfo `json:"players"`
}

// PlayerInfo at time t.
//dont need most of this.
type PlayerInfo struct {
	PlayerSteamID   	int64        `json:"steamID"`
	PlayerName      	string       `json:"name"`
	PlayerTeam      	string       `json:"team"`
	PlayerSide      	string       `json:"side"`
	IsAlive         	bool         `json:"isAlive"`
	EqVal           	int64        `json:"equipmentValue"`
	EqValFreeze     	int64        `json:"equipmentValueFreezetimeEnd"`
	EqValStart      	int64        `json:"equipmentValueRoundStart"`
	Moneytemp			int64        `json:"cashTemp"`
	Money           	int64        `json:"cash"`
	MoneyEarned			int64        `json:"cashEarned"`
	MoneyBeginning		int64        `json:"cashStart"`
	MoneySpentRound 	int64        `json:"cashSpendThisRound"`
	MoneySpentTotal 	int64        `json:"cashSpendTotal"`
}




func convertRoundEndReason(r events.RoundEndReason) string {
	switch reason := r; reason {
	case events.RoundEndReasonTargetBombed:
		return "TargetBombed"
	case events.RoundEndReasonVIPEscaped:
		return "VIPEscaped"
	case events.RoundEndReasonVIPKilled:
		return "VIPKilled"
	case events.RoundEndReasonTerroristsEscaped:
		return "TerroristsEscaped"
	case events.RoundEndReasonCTStoppedEscape:
		return "CTStoppedEscape"
	case events.RoundEndReasonTerroristsStopped:
		return "TerroristsStopped"
	case events.RoundEndReasonBombDefused:
		return "BombDefused"
	case events.RoundEndReasonCTWin:
		return "CTWin"
	case events.RoundEndReasonTerroristsWin:
		return "TerroristsWin"
	case events.RoundEndReasonDraw:
		return "Draw"
	case events.RoundEndReasonHostagesRescued:
		return "HostagesRescued"
	case events.RoundEndReasonTargetSaved:
		return "TargetSaved"
	case events.RoundEndReasonHostagesNotRescued:
		return "HostagesNotRescued"
	case events.RoundEndReasonTerroristsNotEscaped:
		return "TerroristsNotEscaped"
	case events.RoundEndReasonVIPNotEscaped:
		return "VIPNotEscaped"
	case events.RoundEndReasonGameStart:
		return "GameStart"
	case events.RoundEndReasonTerroristsSurrender:
		return "TerroristsSurrender"
	case events.RoundEndReasonCTSurrender:
		return "CTSurrender"
	default:
		return unknown
	}
}






func playerInList(p *common.Player, players []PlayerInfo) bool {
	if len(players) > 0 {
		for _, i := range players {
			if int64(p.SteamID64) == i.PlayerSteamID {
				return true
			}
		}
	}

	return false
}


func parseEconomyduringFT(p *common.Player, currentGame *Game) {

	foundPlayer := false
	if *currentGame.PlayerEconomy != nil {
		for _, i := range *currentGame.PlayerEconomy {
			if int64(p.SteamID64) == i.SteamID {
				
				foundPlayer = true

				if i.MoneyStart == nil {
					i.MoneyStart = int64(p.Money())
				}

				
				i.Money = int64(p.Money())
				i.Moneytemp = int64(p.Money())
				i.MoneySpent = i.MoneyStart - i.Money
				i.MoneyFreeze = int64(p.Money())
			}
		}
	}

	if !foundPlayer {
		currentEconomy := PlayerEconomy{}

		currentEconomy.SteamID = int64(p.SteamID64)
		currentEconomy.Name = p.Name
		currentEconomy.Team = p.TeamState.ClanName()
		currentEconomy.Side = p.TeamState.Side().String()
		currentEconomy.MoneyStart = int64(p.Money())
		currentEconomy.MoneyFreeze = int64(p.Money())
		currentEconomy.Money = int64(p.Money())
		currentEconomy.Moneytemp = int64(p.Money())
		currentEconomy.MoneySpent = currentEconomy.MoneyStart - currentEconomy.Money

		*currentGame.PlayerEconomy = append(*currentGame.PlayerEconomy, currentEconomy)
	}

}



func parseEconomyduringGame(p *common.Player, currentGame *Game){

	
	foundPlayer := false
	if *currentGame.PlayerEconomy != nil {
		for _, i := range *currentGame.PlayerEconomy {
			if int64(p.SteamID64) == i.SteamID {
				
				foundPlayer = true

				if i.MoneyStart == nil {
					i.MoneyStart = int64(p.Money()) + int64(p.MoneySpentThisRound())
				}
				if i.MoneyFreeze == nil {
					i.MoneyFreeze = int64(p.Money())
					i.MoneySpent = i.MoneyStart - i.Money
					i.Moneytemp = int64(p.Money())
				}
				if i.EqValFreeze == nil { i.EqValFreeze = int64(p.EquipmentValueFreezeTimeEnd()) }
				if i.EqValStart == nil { i.EqValStart = int64(p.EquipmentValueRoundStart()) }
				
				i.Money = int64(p.Money())
				i.MoneyEarned = i.MoneyEarned + (i.Money - i.Moneytemp)

				if i.Money - i.Moneytemp > 1300 {
					//setting team reward
				}

				i.Moneytemp = int64(p.Money())

				if p.IsAlive() {
					i.EqVal = int64(p.EquipmentValueCurrent())
				} else { i.EqVal = 0 }
			}
		}
	}

	if !foundPlayer {
		currentEconomy := PlayerEconomy{}

		currentEconomy.SteamID = int64(p.SteamID64)
		currentEconomy.Name = p.Name
		currentEconomy.Team = p.TeamState.ClanName()
		currentEconomy.Side = p.TeamState.Side().String()
		currentEconomy.MoneyStart = int64(p.Money()) + int64(p.MoneySpentThisRound())
		currentEconomy.MoneyFreeze = int64(p.Money())
		currentEconomy.Money = int64(p.Money())
		currentEconomy.Moneytemp = int64(p.Money())
		currentEconomy.MoneySpent = int64(p.MoneySpentThisRound())

		i.EqValFreeze = int64(p.EquipmentValueFreezeTimeEnd()) 
		i.EqValStart = int64(p.EquipmentValueRoundStart()) 

		if p.IsAlive() {
			i.EqVal = int64(p.EquipmentValueCurrent())
		} else { i.EqVal = 0 }

		*currentGame.PlayerEconomy = append(*currentGame.PlayerEconomy, currentEconomy)
	}


	// if Cash start not set, do so right now from last round
	//if FT cash not set, do so right now from current funds

	//if player dead, set EqVal 0

	
	
}




func parsePlayer(gs dem.GameState, p *common.Player) PlayerInfo {
	currentPlayer := PlayerInfo{}
	currentPlayer.PlayerSteamID = int64(p.SteamID64)
	currentPlayer.PlayerName = p.Name
	if p.TeamState != nil {
		currentPlayer.PlayerTeam = p.TeamState.ClanName()
	}

	switch p.Team {
	case common.TeamTerrorists:
		currentPlayer.PlayerSide = "T"
	case common.TeamCounterTerrorists:
		currentPlayer.PlayerSide = "CT"
	case common.TeamSpectators:
		currentPlayer.PlayerSide = spectator
	case common.TeamUnassigned:
		currentPlayer.PlayerSide = unassigned
	default:
		currentPlayer.PlayerSide = unknown
	}

	//TODO: this is where I can get the other metrics I want 

	if currentPlayer.MoneyBeginning == nil {
		currentPlayer.MoneyBeginning = int64(p.Money())
	}

	if 

	currentPlayer.Moneytemp = currentPlayer.Money

	// Calc other metrics
	currentPlayer.IsAlive = p.IsAlive()
	currentPlayer.Money = int64(p.Money())
	currentPlayer.MoneySpentRound = int64(p.MoneySpentThisRound())
	currentPlayer.MoneySpentTotal = int64(p.MoneySpentTotal())
	currentPlayer.EqVal = int64(p.EquipmentValueCurrent())
	currentPlayer.EqValFreeze = int64(p.EquipmentValueFreezeTimeEnd())
	currentPlayer.EqValStart = int64(p.EquipmentValueRoundStart())




	return currentPlayer
}



func countAlivePlayers(players []PlayerInfo) int64 {
	var alivePlayers int64
	for _, p := range players {
		if p.IsAlive {
			alivePlayers++
		}
	}

	return alivePlayers
}



// Define cleaning functions.
func cleanMapName(mapName string) string {
	lastSlash := strings.LastIndex(mapName, "/")
	if lastSlash == -1 {
		return mapName
	}

	return mapName[lastSlash+1:]
}


func appendFrameToRound(currentRound *GameRound, currentFrame *GameFrame, globalFrameIndex *int64) {
	currentFrame.FrameID = int64(len(currentRound.Frames))
	currentFrame.GlobalFrameID = *globalFrameIndex
	*globalFrameIndex++
	currentRound.Frames = append(currentRound.Frames, *currentFrame)
}


func initializeRound(currentRound *GameRound) {
	currentRound.PlayerEconomy = []PlayerEconomy{}
	currentRound.Frames = []GameFrame{}
}


func registerConnectHandler(demoParser *dem.Parser, currentGame *Game) {
	(*demoParser).RegisterEventHandler(func(e events.PlayerConnect) {
		if e.Player != nil {
			gs := (*demoParser).GameState()
			playerConnected := ConnectAction{}

			playerConnected.Tick = int64(gs.IngameTick())
			playerConnected.ConnectType = "connect"
			playerConnected.SteamID = e.Player.SteamID64

			currentGame.Connections = append(currentGame.Connections, playerConnected)
		}
	})
}
func registerDisconnectHandler(demoParser *dem.Parser, currentGame *Game) {
	(*demoParser).RegisterEventHandler(func(e events.PlayerDisconnected) {
		if e.Player != nil {
			gs := (*demoParser).GameState()
			playerConnected := ConnectAction{}

			playerConnected.Tick = int64(gs.IngameTick())
			playerConnected.ConnectType = "disconnect"
			playerConnected.SteamID = e.Player.SteamID64

			currentGame.Connections = append(currentGame.Connections, playerConnected)
		}
	})
}

//do I need this?
func registerMatchphases(demoParser *dem.Parser, currentGame *Game) {
	(*demoParser).RegisterEventHandler(func(e events.AnnouncementLastRoundHalf) {
		gs := (*demoParser).GameState()

		currentGame.MatchPhases.AnnLastRoundHalf = append(currentGame.MatchPhases.AnnLastRoundHalf,
			int64(gs.IngameTick()))
	})

	(*demoParser).RegisterEventHandler(func(e events.AnnouncementFinalRound) {
		gs := (*demoParser).GameState()

		currentGame.MatchPhases.AnnFinalRound = append(currentGame.MatchPhases.AnnFinalRound,
			int64(gs.IngameTick()))
	})

	(*demoParser).RegisterEventHandler(func(e events.AnnouncementMatchStarted) {
		gs := (*demoParser).GameState()

		currentGame.MatchPhases.AnnMatchStarted = append(currentGame.MatchPhases.AnnMatchStarted,
			int64(gs.IngameTick()))
	})

	(*demoParser).RegisterEventHandler(func(e events.GameHalfEnded) {
		gs := (*demoParser).GameState()

		currentGame.MatchPhases.GameHalfEnded = append(currentGame.MatchPhases.GameHalfEnded,
			int64(gs.IngameTick()))
	})

	(*demoParser).RegisterEventHandler(func(e events.MatchStart) {
		gs := (*demoParser).GameState()

		currentGame.MatchPhases.MatchStart = append(currentGame.MatchPhases.MatchStart,
			int64(gs.IngameTick()))
	})

	(*demoParser).RegisterEventHandler(func(e events.MatchStartedChanged) {
		gs := (*demoParser).GameState()

		currentGame.MatchPhases.MatchStartedChanged = append(currentGame.MatchPhases.MatchStartedChanged,
			int64(gs.IngameTick()))
	})

	(*demoParser).RegisterEventHandler(func(e events.IsWarmupPeriodChanged) {
		gs := (*demoParser).GameState()

		currentGame.MatchPhases.WarmupChanged = append(currentGame.MatchPhases.WarmupChanged,
			int64(gs.IngameTick()))
	})

	(*demoParser).RegisterEventHandler(func(e events.TeamSideSwitch) {
		gs := (*demoParser).GameState()

		currentGame.MatchPhases.TeamSwitch = append(currentGame.MatchPhases.TeamSwitch,
			int64(gs.IngameTick()))
	})
}

func setTeamValuesInRound(currentRound *GameRound, gameState *dem.GameState) {
	if (*gameState).TeamTerrorists() != nil {
		currentRound.TScore = int64((*gameState).TeamTerrorists().Score())
		tTeam := (*gameState).TeamTerrorists().ClanName()
		currentRound.TTeam = &tTeam
	}
	if (*gameState).TeamCounterTerrorists() != nil {
		currentRound.CTScore = int64((*gameState).TeamCounterTerrorists().Score())
		ctTeam := (*gameState).TeamCounterTerrorists().ClanName()
		currentRound.CTTeam = &ctTeam
	}
}


func registerRoundStartHandler(demoParser *dem.Parser, currentGame *Game, currentRound *GameRound,
	roundStarted *int, roundInFreezetime *int, roundInEndTime *int, freezeTimeSetFromGameState *bool,
	globalFrameIndex *int64) {
	(*demoParser).RegisterEventHandler(func(e events.RoundStart) {
		gs := (*demoParser).GameState()
		currentGame.MatchPhases.RoundStarted = append(currentGame.MatchPhases.RoundStarted, int64(gs.IngameTick()))

		if *roundStarted == 1 {
			currentGame.Rounds = append(currentGame.Rounds, *currentRound)
		} else {
			*globalFrameIndex = 0
		}

		*roundStarted = 1
		*roundInFreezetime = 1
		*freezeTimeSetFromGameState = false
		*roundInEndTime = 0
		*currentRound = GameRound{}

		*lastKiller = nil
		*lastDefuser = nil
		*lastPlanter = nil
		*roundRewardSet = false
		


		// Create empty action lists
		initializeRound(currentRound)

		// Parse flags
		currentRound.IsWarmup = gs.IsWarmupPeriod()
		currentRound.RoundNum = int64(len(currentGame.Rounds) + 1)
		currentRound.StartTick = int64(gs.IngameTick())

		setTeamValuesInRound(currentRound, &gs)


		parseEconomyatStart()

		// Parse the players
		teamCT := PlayerTeams{}
		if gs.TeamCounterTerrorists() != nil {
			teamCT.TeamName = gs.TeamCounterTerrorists().ClanName()
			for _, player := range gs.TeamCounterTerrorists().Members() {
				pl := Players{}
				pl.PlayerName = player.Name
				pl.SteamID = int64(player.SteamID64)
				foundPlayer := false
				for _, p := range teamCT.Players {
					if p.SteamID == pl.SteamID {
						foundPlayer = true
					}
				}
				if !foundPlayer {
					teamCT.Players = append(teamCT.Players, pl)
				}
			}
		}
		currentRound.CTSide = teamCT

		teamT := PlayerTeams{}
		if gs.TeamTerrorists() != nil {
			teamT.TeamName = gs.TeamTerrorists().ClanName()
			for _, player := range gs.TeamTerrorists().Members() {
				pl := Players{}
				pl.PlayerName = player.Name
				pl.SteamID = int64(player.SteamID64)
				foundPlayer := false
				for _, p := range teamT.Players {
					if p.SteamID == pl.SteamID {
						foundPlayer = true
					}
				}
				if !foundPlayer {
					teamT.Players = append(teamT.Players, pl)
				}
			}
		}
		currentRound.TSide = teamT
	})
}

func registerRoundFreezeTimeEndHandler(demoParser *dem.Parser, currentGame *Game, currentRound *GameRound,
	convParsed *int, roundRestartDelay *int64,
	roundStarted *int, roundInFreezetime *int, roundInEndTime *int, freezeTimeSetFromGameState *bool) {
	// Parse round freezetime ends
	(*demoParser).RegisterEventHandler(func(e events.RoundFreezetimeEnd) {
		gs := (*demoParser).GameState()
		currentGame.MatchPhases.RoundFreezeEnded = append(currentGame.MatchPhases.RoundFreezeEnded,
			int64(gs.IngameTick()))

		// Reupdate the teams to make sure
		setTeamValuesInRound(currentRound, &gs)



		// Determine if round is still in warmup mode
		currentRound.IsWarmup = gs.IsWarmupPeriod()

		// If convars aren't parsed, do so
		if *convParsed == 0 {
			// If convars are unparsed, record the convars of the server
			serverConfig := ServerConVar{}
			conv := gs.Rules().ConVars()
			serverConfig.CashBombDefused, _ = strconv.ParseInt(conv["cash_player_bomb_defused"], 10, 64)
			serverConfig.CashBombPlanted, _ = strconv.ParseInt(conv["cash_player_bomb_planted"], 10, 64)
			serverConfig.CashWinBomb, _ = strconv.ParseInt(conv["cash_team_terrorist_win_bomb"], 10, 64)
			serverConfig.CashWinDefuse, _ = strconv.ParseInt(conv["cash_team_win_by_defusing_bomb"], 10, 64)
			serverConfig.CashWinTimeRunOut, _ = strconv.ParseInt(conv["cash_team_win_by_time_running_out_bomb"], 10, 64)
			serverConfig.CashWinElimination, _ = strconv.ParseInt(conv["cash_team_elimination_bomb_map"], 10, 64)
			serverConfig.CashPlayerKilledDefault, _ = strconv.ParseInt(conv["cash_player_killed_enemy_default"], 10, 64)
			serverConfig.CashTeamLoserBonus, _ = strconv.ParseInt(conv["cash_team_loser_bonus"], 10, 64)
			serverConfig.CashTeamLoserBonusConsecutive, _ = strconv.ParseInt(
				conv["cash_team_loser_bonus_consecutive_rounds"], 10, 64)
			serverConfig.MaxRounds, _ = strconv.ParseInt(conv["mp_maxrounds"], 10, 64)
			serverConfig.RoundTime, _ = strconv.ParseInt(conv["mp_roundtime"], 10, 64)
			serverConfig.RoundTimeDefuse, _ = strconv.ParseInt(conv["mp_roundtime_defuse"], 10, 64)
			serverConfig.RoundRestartDelay, _ = strconv.ParseInt(conv["mp_round_restart_delay"], 10, 64)
			serverConfig.FreezeTime, _ = strconv.ParseInt(conv["mp_freezetime"], 10, 64)
			serverConfig.BuyTime, _ = strconv.ParseInt(conv["mp_buytime"], 10, 64)
			serverConfig.BombTimer, _ = strconv.ParseInt(conv["mp_c4timer"], 10, 64)
			serverConfig.TimeoutsAllowed, _ = strconv.ParseInt(conv["mp_team_timeout_max"], 10, 64)
			serverConfig.CoachingAllowed, _ = strconv.ParseInt(conv["sv_coaching_enabled"], 10, 64)
			currentGame.ServerVars = serverConfig
			*convParsed = 1

			// Change so that round restarts are parsed using the server convar
			if serverConfig.RoundRestartDelay == 0 {
				*roundRestartDelay = 5 // This is default on many servers, I think
			} else {
				*roundRestartDelay = serverConfig.RoundRestartDelay
			}
		}

		if *roundInFreezetime == 0 && !(*freezeTimeSetFromGameState) {
			// This means the RoundStart event did not fire, but the FreezeTimeEnd did
			currentGame.Rounds = append(currentGame.Rounds, *currentRound)
			*roundStarted = 1
			*roundInEndTime = 0
			*currentRound = GameRound{}

			*lastKiller = nil
			*lastDefuser = nil
			*lastPlanter = nil
			*roundRewardSet = false

			// Create empty action lists
			initializeRound(currentRound)

			currentRound.IsWarmup = gs.IsWarmupPeriod()
			currentRound.RoundNum = int64(len(currentGame.Rounds) + 1)
			currentRound.StartTick = int64(gs.IngameTick() - int(currentGame.TickRate)*int(currentGame.ServerVars.FreezeTime))
			currentRound.FreezeTimeEndTick = int64(gs.IngameTick())
			setTeamValuesInRound(currentRound, &gs)
		}

		// Parse the players
		teamCT := PlayerTeams{}
		if gs.TeamCounterTerrorists() != nil {
			teamCT.TeamName = gs.TeamCounterTerrorists().ClanName()
			for _, player := range gs.TeamCounterTerrorists().Members() {
				pl := Players{}
				pl.PlayerName = player.Name
				pl.SteamID = int64(player.SteamID64)
				foundPlayer := false
				for _, p := range teamCT.Players {
					if p.SteamID == pl.SteamID {
						foundPlayer = true
					}
				}
				if !foundPlayer {
					teamCT.Players = append(teamCT.Players, pl)
				}
			}
		}
		currentRound.CTSide = teamCT

		teamT := PlayerTeams{}
		if gs.TeamTerrorists() != nil {
			teamT.TeamName = gs.TeamTerrorists().ClanName()
			for _, player := range gs.TeamTerrorists().Members() {
				pl := Players{}
				pl.PlayerName = player.Name
				pl.SteamID = int64(player.SteamID64)
				foundPlayer := false
				for _, p := range teamT.Players {
					if p.SteamID == pl.SteamID {
						foundPlayer = true
					}
				}
				if !foundPlayer {
					teamT.Players = append(teamT.Players, pl)
				}
			}
		}
		currentRound.TSide = teamT

		*roundInFreezetime = 0
		*freezeTimeSetFromGameState = false
		currentRound.FreezeTimeEndTick = int64(gs.IngameTick())
	})
}

func registerRoundEndOfficialHandler(demoParser *dem.Parser, currentGame *Game, currentRound *GameRound,
	roundInEndTime *int, roundRestartDelay *int64) {
	(*demoParser).RegisterEventHandler(func(e events.RoundEndOfficial) {
		gs := (*demoParser).GameState()
		currentGame.MatchPhases.RoundEndedOfficial = append(currentGame.MatchPhases.RoundEndedOfficial,
			int64(gs.IngameTick()))


		//getting round reward somewhere here and in roundend register:
		//if !kill and !defuser and !planter and round end and round reward not set yet and difference in money 
		// (edgecase:where bomb gets defused and 2 different players get kills and only afterwards the round end gets triggered,
		// the first killer should not be used for round reward)
		// , use this player 
		
		//dont forget to parse players


		if *roundInEndTime == 0 {
			currentRound.EndTick = int64(gs.IngameTick()) - (*roundRestartDelay * currentGame.TickRate)
			currentRound.EndOfficialTick = int64(gs.IngameTick())

			

			// Parse who won the round, not great...but a stopgap measure
			aliveT := 0
			if gs.TeamTerrorists() != nil {
				tPlayers := gs.TeamTerrorists().Members()
				for _, p := range tPlayers {
					if p != nil && p.IsAlive() {
						aliveT++
					}
				}
			}

			aliveCT := 0
			if gs.TeamCounterTerrorists() != nil {
				ctPlayers := gs.TeamCounterTerrorists().Members()
				for _, p := range ctPlayers {
					if p != nil && p.IsAlive() {
						aliveCT++
					}
				}
			}

			if aliveCT == 0 {
				currentRound.Reason = "TerroristsWin"
				currentRound.EndTScore = currentRound.TScore + 1
				currentRound.EndCTScore = currentRound.CTScore
				if (gs.TeamTerrorists() != nil) && (gs.TeamCounterTerrorists() != nil) {
					tTeam := gs.TeamTerrorists().ClanName()
					ctTeam := gs.TeamCounterTerrorists().ClanName()
					currentRound.WinningTeam = &tTeam
					currentRound.LosingTeam = &ctTeam
				}
				currentRound.WinningSide = "T"
			} else {
				currentRound.Reason = "CTWin"
				currentRound.EndCTScore = currentRound.CTScore + 1
				currentRound.EndTScore = currentRound.TScore
				if (gs.TeamTerrorists() != nil) && (gs.TeamCounterTerrorists() != nil) {
					tTeam := gs.TeamTerrorists().ClanName()
					ctTeam := gs.TeamCounterTerrorists().ClanName()
					currentRound.WinningTeam = &ctTeam
					currentRound.LosingTeam = &tTeam
				}
				currentRound.WinningSide = "CT"
			}
		} else {
			currentRound.EndOfficialTick = int64(gs.IngameTick())
		}
	})
}

func registerRoundEndHandler(demoParser *dem.Parser, currentGame *Game, currentRound *GameRound,
	roundStarted *int, roundInEndTime *int, roundRestartDelay *int64) {
	(*demoParser).RegisterEventHandler(func(e events.RoundEnd) {
		gs := (*demoParser).GameState()

		//getting round reward somewhere here and in roundend offical  handler:
		//if !kill and !defuser and !planter and round end and round reward not set yet, use this player 

		//dont forget to parse players

		if *roundStarted == 1 {
			if (gs.TeamTerrorists() != nil) && (gs.TeamCounterTerrorists() != nil) {
				tTeam := gs.TeamTerrorists().ClanName()
				ctTeam := gs.TeamCounterTerrorists().ClanName()
				currentRound.TTeam = &tTeam
				currentRound.CTTeam = &ctTeam
			}
		}

		currentGame.MatchPhases.RoundEnded = append(currentGame.MatchPhases.RoundEnded, int64(gs.IngameTick()))

		if *roundStarted == 0 {
			*roundStarted = 1

			currentRound.RoundNum = 0
			currentRound.StartTick = 0
			currentRound.TScore = 0
			currentRound.CTScore = 0
			if (gs.TeamTerrorists() != nil) && (gs.TeamCounterTerrorists() != nil) {
				tTeam := gs.TeamTerrorists().ClanName()
				ctTeam := gs.TeamCounterTerrorists().ClanName()
				currentRound.TTeam = &tTeam
				currentRound.CTTeam = &ctTeam
			}
		}

		*roundInEndTime = 1

		var winningTeam string
		switch e.Winner {
		case common.TeamTerrorists:
			winningTeam = "T"
			currentRound.EndTScore = currentRound.TScore + 1
			currentRound.EndCTScore = currentRound.CTScore
		case common.TeamCounterTerrorists:
			winningTeam = "CT"
			currentRound.EndCTScore = currentRound.CTScore + 1
			currentRound.EndTScore = currentRound.TScore
		case common.TeamSpectators:
			winningTeam = "Spectators"
		case common.TeamUnassigned:
			winningTeam = unassigned
		default:
			winningTeam = unknown
		}

		currentRound.EndTick = int64(gs.IngameTick())
		currentRound.EndOfficialTick = int64(gs.IngameTick()) + (*roundRestartDelay * currentGame.TickRate)
		currentRound.Reason = convertRoundEndReason(e.Reason)
		currentRound.WinningSide = winningTeam

		if (gs.TeamTerrorists() != nil) && (gs.TeamCounterTerrorists() != nil) {
			tTeam := gs.TeamTerrorists().ClanName()
			ctTeam := gs.TeamCounterTerrorists().ClanName()

			if winningTeam == "CT" {
				currentRound.LosingTeam = &tTeam
				currentRound.WinningTeam = &ctTeam
			} else if winningTeam == "T" {
				currentRound.LosingTeam = &ctTeam
				currentRound.WinningTeam = &tTeam
			}
		}


	})
}


func inFreezeTimeFromGameRules(gameState *dem.GameState) bool {
	entity := (*gameState).Rules().Entity()
	if entity != nil {
		property, found := entity.PropertyValue("cs_gamerules_data.m_bFreezePeriod")
		if found {
			return property.BoolVal()
		}
	}

	return false
}


//probably needs to be reduced, if not completly emptied.
//todo: check how the parserates are used
func registerFrameHandler(demoParser *dem.Parser, currentGame *Game, currentRound *GameRound, 
	roundInFreezetime *int, roundInEndTime *int, freezeTimeSetFromGameState *bool,
	currentFrameIdx *int, parseFrames *bool, globalFrameIndex *int64) {
	(*demoParser).RegisterEventHandler(func(e events.FrameDone) {
		gs := (*demoParser).GameState()

		// If the game says we are not in freeze time anymore
		// but the toggle still thinks we are then correct the toggle
		if !inFreezeTimeFromGameRules(&gs) && (*roundInFreezetime != 0) {
			*roundInFreezetime = 0
			*freezeTimeSetFromGameState = true
			currentRound.FreezeTimeEndTick = int64(gs.IngameTick())
		}

		if (*roundInFreezetime == 0) && (*roundInEndTime == 0) {
			if gs.TeamCounterTerrorists() != nil {
				currentRound.CTRoundStartEqVal = int64(gs.TeamCounterTerrorists().RoundStartEquipmentValue())
				currentRound.CTFreezeTimeEndEqVal = int64(gs.TeamCounterTerrorists().FreezeTimeEndEquipmentValue())
			}
			if gs.TeamTerrorists() != nil {
				currentRound.TRoundStartEqVal = int64(gs.TeamTerrorists().RoundStartEquipmentValue())
				currentRound.TFreezeTimeEndEqVal = int64(gs.TeamTerrorists().FreezeTimeEndEquipmentValue())
			}
		}

		if (*roundInFreezetime == 0) && (*currentFrameIdx == 0) && *parseFrames {
			currentFrame := GameFrame{}
			currentFrame.IsKillFrame = false

			// Create empty player lists
			currentFrame.CT.Players = []PlayerInfo{}
			currentFrame.T.Players = []PlayerInfo{}

			currentFrame.Tick = int64(gs.IngameTick())


			// Parse T
			currentFrame.T = TeamFrameInfo{}
			currentFrame.T.Side = "T"
			if gs.TeamTerrorists() != nil {
				currentFrame.T.Team = gs.TeamTerrorists().ClanName()
				currentFrame.T.CurrentEqVal = int64(gs.TeamTerrorists().CurrentEquipmentValue())
				tPlayers := gs.TeamTerrorists().Members()

				for _, p := range tPlayers {
					if p != nil {
						if !playerInList(p, currentFrame.T.Players) {
							currentFrame.T.Players = append(currentFrame.T.Players, parsePlayer(gs, p))
						}
					}
				}
			}

			currentFrame.T.AlivePlayers = countAlivePlayers(currentFrame.T.Players)
			currentFrame.T.TotalUtility = countUtility(currentFrame.T.Players)
			// currentFrame.T.CurrentEqVal = sumPlayerEqVal(currentFrame.T.Players)

			// Parse CT
			currentFrame.CT = TeamFrameInfo{}
			currentFrame.CT.Side = "CT"
			if gs.TeamCounterTerrorists() != nil {
				currentFrame.CT.Team = gs.TeamCounterTerrorists().ClanName()
				currentFrame.CT.CurrentEqVal = int64(gs.TeamCounterTerrorists().CurrentEquipmentValue())
				ctPlayers := gs.TeamCounterTerrorists().Members()

				for _, p := range ctPlayers {
					if p != nil {
						if !playerInList(p, currentFrame.CT.Players) {
							currentFrame.CT.Players = append(currentFrame.CT.Players, parsePlayer(gs, p))
						}
					}
				}
			}

			currentFrame.CT.AlivePlayers = countAlivePlayers(currentFrame.CT.Players)
			currentFrame.CT.TotalUtility = countUtility(currentFrame.CT.Players)
			// currentFrame.CT.CurrentEqVal = sumPlayerEqVal(currentFrame.CT.Players)

			
			




			// Add frame
			if (len(currentFrame.CT.Players) > 0) || (len(currentFrame.T.Players) > 0) {
				if len(currentRound.Frames) > 0 {
					if currentRound.Frames[len(currentRound.Frames)-1].Tick < currentFrame.Tick {
						appendFrameToRound(currentRound, &currentFrame, globalFrameIndex)
					}
				} else {
					appendFrameToRound(currentRound, &currentFrame, globalFrameIndex)
				}
			}
		}

		//this is where parserate plays a role. so every thing gets parsed but if parse rate is high
		//the entire frame does not get recorded. It can be relevant for me to correctly record
		//the cash and equipment value of the players. I could manually set a very high parse rate
		//or completely ignore frames unless I need them for cleaning purposes.
		if *currentFrameIdx == (currentGame.ParsingOpts.ParseRate - 1) {
			*currentFrameIdx = 0
		} else {
			*currentFrameIdx++
		}
	})
}

//additional cleaning function is necessary -> see python code
func cleanAndWriteGame(currentGame *Game, jsonIndentation bool, outpath string) {
	// Loop through damages and see if there are any multi-damages in a single tick,
	// and reduce them to one attacker-victim-weapon entry per tick
	if currentGame.ParsingOpts.DamagesRolled {
		for i := range currentGame.Rounds {
			var tempDamages []DamageAction
			for j := range currentGame.Rounds[i].Damages {
				if j < len(currentGame.Rounds[i].Damages) && j > 0 {
					if (len(tempDamages) > 0) &&
						(currentGame.Rounds[i].Damages[j].Tick == tempDamages[len(tempDamages)-1].Tick) &&
						(currentGame.Rounds[i].Damages[j].AttackerSteamID == tempDamages[len(tempDamages)-1].AttackerSteamID) &&
						(currentGame.Rounds[i].Damages[j].VictimSteamID == tempDamages[len(tempDamages)-1].VictimSteamID) &&
						(currentGame.Rounds[i].Damages[j].Weapon == tempDamages[len(tempDamages)-1].Weapon) {
						tempDamages[len(tempDamages)-1].HpDamage += currentGame.Rounds[i].Damages[j].HpDamage
						tempDamages[len(tempDamages)-1].HpDamageTaken += currentGame.Rounds[i].Damages[j].HpDamageTaken
						tempDamages[len(tempDamages)-1].ArmorDamage += currentGame.Rounds[i].Damages[j].ArmorDamage
						tempDamages[len(tempDamages)-1].ArmorDamageTaken += currentGame.Rounds[i].Damages[j].ArmorDamageTaken
					} else {
						tempDamages = append(tempDamages, currentGame.Rounds[i].Damages[j])
					}
					tempDamages = append(tempDamages, currentGame.Rounds[i].Damages[j])
				}
			}
			currentGame.Rounds[i].Damages = tempDamages
		}
	}

	// Write the JSON
	var file []byte
	var err error
	if jsonIndentation {
		file, err = json.MarshalIndent(currentGame, "", " ")
	} else {
		file, err = json.Marshal(currentGame)
	}
	checkError(err)
	_ = os.WriteFile(outpath+"/"+currentGame.MatchName+".json", file, 0600)
}

// Main.
func main() {
	/* Parse the arguments

	Run the parser as follows:
	go run parse_demo.go -demo /path/to/demo.dem -parserate 1/2/4/8/16/32/64/128 -demoID someDemoIDString

	The parserate should be one of 2^0 to 2^7. The lower the value, the more frames are collected.
	Indicates spacing between parsed demo frames in ticks.
	*/
	logger := log.New(os.Stderr, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)

	fl := new(flag.FlagSet)
	demoPathPtr := fl.String("demo", "", "Demo file `path`")
	parseRatePtr := fl.Int("parserate", 128, "Parse rate, indicates spacing between ticks")
	parseFramesPtr := fl.Bool("parseframes", false, "Parse frames")
	parseKillFramesPtr := fl.Bool("parsekillframes", false, "Parse kill frames")
	tradeTimePtr := fl.Int("tradetime", 5, "Trade time frame (in seconds)")
	roundBuyPtr := fl.String("buystyle", "hltv", "Round buy style")
	damagesRolledPtr := fl.Bool("dmgrolled", false, "Roll up damages")
	demoIDPtr := fl.String("demoid", "", "Demo string ID")
	jsonIndentationPtr := fl.Bool("jsonindentation", false, "Indent JSON file")
	parseChatPtr := fl.Bool("parsechat", false, "Parse chat messages")
	outpathPtr := fl.String("out", "", "Path to write output JSON")

	err := fl.Parse(os.Args[1:])
	checkError(err)

	demPath := *demoPathPtr
	parseRate := *parseRatePtr
	parseFrames := *parseFramesPtr
	parseKillFrames := *parseKillFramesPtr
	tradeTime := int64(*tradeTimePtr)
	roundBuyStyle := *roundBuyPtr
	damagesRolled := *damagesRolledPtr
	jsonIndentation := *jsonIndentationPtr
	parseChat := *parseChatPtr
	outpath := *outpathPtr

	// Read in demofile
	f, err := os.Open(demPath)
	checkError(err)
	defer f.Close()

	// Create new demoparser
	p := dem.NewParser(f)
	defer p.Close()

	// Parse demofile header
	header, err := p.ParseHeader()
	checkError(err)

	// Parse nav mesh given the map name
	currentMap := header.MapName
	currentMap = cleanMapName(currentMap)

	// Create flags to guide parsing
	roundStarted := 0
	roundInEndTime := 0
	roundInFreezetime := 0
	freezeTimeSetFromGameState := false
	currentFrameIdx := 0
	convParsed := 0

	// Create game object, then initial round object
	currentGame := Game{}
	currentGame.MatchName = *demoIDPtr
	currentGame.Map = cleanMapName(currentMap)
	if p.TickRate() == 0 {
		currentGame.TickRate = 128
	} else {
		currentGame.TickRate = int64(math.Round(p.TickRate())) // Rounds to 127 instead
	}
	currentGame.PlaybackTicks = int64(header.PlaybackTicks)
	currentGame.PlaybackFrames = int64(header.PlaybackFrames)
	currentGame.ClientName = header.ClientName

	// Set parsing options
	parsingOpts := ParserOpts{}
	parsingOpts.ParseRate = parseRate
	parsingOpts.ParseFrames = parseFrames
	parsingOpts.ParseKillFrames = parseKillFrames
	parsingOpts.TradeTime = tradeTime
	parsingOpts.RoundBuyStyle = roundBuyStyle
	parsingOpts.DamagesRolled = damagesRolled
	parsingOpts.ParseChat = parseChat
	currentGame.ParsingOpts = parsingOpts

	globalFrameIndex := int64(0)

	currentRound := GameRound{}

	// Create empty action lists for first round
	initializeRound(&currentRound)

	RoundRestartDelay := int64(5)

	// Create empty lists
	currentGame.MatchPhases.AnnFinalRound = []int64{}
	currentGame.MatchPhases.AnnLastRoundHalf = []int64{}
	currentGame.MatchPhases.AnnMatchStarted = []int64{}
	currentGame.MatchPhases.GameHalfEnded = []int64{}
	currentGame.MatchPhases.MatchStart = []int64{}
	currentGame.MatchPhases.MatchStartedChanged = []int64{}
	currentGame.MatchPhases.WarmupChanged = []int64{}
	currentGame.MatchPhases.TeamSwitch = []int64{}
	currentGame.MatchPhases.RoundStarted = []int64{}
	currentGame.MatchPhases.RoundFreezeEnded = []int64{}
	currentGame.MatchPhases.RoundEnded = []int64{}
	currentGame.MatchPhases.RoundEndedOfficial = []int64{}



	// Parse player connects
	registerConnectHandler(&p, &currentGame)

	// Parse player disconnects
	registerDisconnectHandler(&p, &currentGame)

	// Parse the match phases
	registerMatchphases(&p, &currentGame)


	// Parse round starts
	registerRoundStartHandler(&p, &currentGame, &currentRound,
		&roundStarted, &roundInFreezetime, &roundInEndTime, &freezeTimeSetFromGameState,
		, &globalFrameIndex)
	registerRoundFreezeTimeEndHandler(&p, &currentGame, &currentRound, &convParsed,
		&RoundRestartDelay, &roundStarted, &roundInFreezetime, &roundInEndTime, &freezeTimeSetFromGameState)

	// Parse round ends
	registerRoundEndOfficialHandler(&p, &currentGame, &currentRound, &roundInEndTime, &RoundRestartDelay)
	registerRoundEndHandler(&p, &currentGame, &currentRound, &roundStarted, &roundInEndTime, &RoundRestartDelay)
	
	// Parse a demo frame. If parse rate is 1, then every frame is parsed.
	// If parse rate is 2, then every 2 frames is parsed, and so on
	
	registerFrameHandler(&p, &currentGame, &currentRound,  &roundInFreezetime,
		&roundInEndTime, &freezeTimeSetFromGameState, &currentFrameIdx, &parseFrames, &globalFrameIndex)

	// Parse demofile to end
	err = p.ParseToEnd()
	currentGame.ParsedToFrame = int64(p.CurrentFrame())

	// Add the most recent round
	currentGame.Rounds = append(currentGame.Rounds, currentRound)

	// Clean rounds
	if len(currentGame.Rounds) > 0 {
		cleanAndWriteGame(&currentGame, jsonIndentation, outpath)
	}

	// Check error
	if err != nil {
		if errors.Is(err, dem.ErrUnexpectedEndOfDemo) {
			logger.Println(err)
			logger.Println("ErrUnexpectedEndOfDemo signals that the demo" +
				" is incomplete / corrupt - these demos may still be useful," +
				" check how far the parser got.")
		}
	}
}

// Function to handle errors.
func checkError(err error) {
	if err != nil {
		panic(err)
	}
}
