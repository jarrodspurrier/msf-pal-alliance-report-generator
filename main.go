package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/sheets/v4"
)

// MSFCharacter contains the details of a character in MSF.
type MSFCharacter struct {
	ID        string `json:"id"`
	MsfGgID   string `json:"msf.gg.id"`
	MsfGgName string `json:"msf.gg.name"`
	Avatar    string `json:"avatar"`
	Labels    struct {
		En string `json:"en"`
		Fr string `json:"fr"`
	} `json:"labels"`
	Traits     []string `json:"traits"`
	BlitzRoles []string `json:"blitzRoles,omitempty"`
	Speed      int      `json:"speed"`
	Synergies  []struct {
		Capacity string `json:"capacity"`
		Min      int    `json:"min"`
	} `json:"synergies,omitempty"`
}

// MSFCharacters is a list of MSF characters.
type MSFCharacters []MSFCharacter

// MSFPlayerCharacter contains the current state of a player's MSF character.
type MSFPlayerCharacter struct {
	Basic       int    `json:"basic"`
	Favorite    bool   `json:"favorite"`
	GearLevel   int    `json:"gearLevel"`
	ID          string `json:"id"`
	Level       int    `json:"level"`
	Passive     int    `json:"passive"`
	Player      string `json:"player"`
	Power       int    `json:"power"`
	RedStars    int    `json:"redStars"`
	Special     int    `json:"special"`
	Ultimate    int    `json:"ultimate"`
	Unlocked    bool   `json:"unlocked"`
	YellowStars int    `json:"yellowStars"`
}

// MSFPlayerCharacters is a list containing a player's MSF characters.
type MSFPlayerCharacters []MSFPlayerCharacter

// MSFTeam defines details of a MSF team and the characters in them.
type MSFTeam struct {
	Name       string
	Label      string
	Characters []string
}

// ByAverageTotalTeamPower sorts cell values by the player with the highest average team power in descending order.
type ByAverageTotalTeamPower [][]interface{}

func (a ByAverageTotalTeamPower) Len() int      { return len(a) }
func (a ByAverageTotalTeamPower) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByAverageTotalTeamPower) Less(i, j int) bool {
	averageTotalTeamPowerA, ok := a[i][len(a[i])-1].(int)
	if !ok {
		return true
	}

	averageTotalTeamPowerB, ok := a[j][len(a[j])-1].(int)
	if !ok {
		return false
	}

	return averageTotalTeamPowerA > averageTotalTeamPowerB
}

const (
	msfAllianceID = "10112006-6d29-48d9-8877-9faf91df83d9"
	msfPalAPIKey  = "8c1cd6fc-7e90-467d-ba63-4a6492dc018f"
	spreadsheetID = "1p-nLLUjPpkNGiMunDCKryYgRtff8h_P_ag8nK7w7FkI"
)

var (
	sheetsService *sheets.Service

	sheetCellIndexToLetter = []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z"}

	aim = MSFTeam{
		Name:  "A.I.M.",
		Label: "AIM",
		Characters: []string{
			"scientist-supreme",
			"graviton",
			"aim-monstruosity",
			"aim-assualter",
			"aim-security",
		},
	}

	asgardian = MSFTeam{
		Name:  "Asgardian",
		Label: "ASG",
		Characters: []string{
			"hela",
			"thor",
			"loki",
			"sif",
			"heimdall",
		},
	}

	avengers = MSFTeam{
		Name:  "Avengers",
		Label: "AVG",
		Characters: []string{
			"captain-america",
			"captain-marvel",
			"hulk",
			"black-widow",
			"hawkeye",
		},
	}

	blackOrder = MSFTeam{
		Name:  "Black Order",
		Label: "BO",
		Characters: []string{
			"ebony-maw",
			"thanos",
			"cull-obsidian",
			"corvus-glaive",
			"proxima-midnight",
		},
	}

	brawlers = MSFTeam{
		Name:  "Brawlers",
		Label: "BRAWL",
		Characters: []string{
			"america-chavez",
			"squirrelgirl",
			"miss-marvel",
			"psylocke",
			"wolverine",
		},
	}

	brotherhoodV2 = MSFTeam{
		Name:  "Brotherhood V2",
		Label: "BH2",
		Characters: []string{
			"juggernaut",
			"toad",
			"blob",
			"pyro",
			"magneto",
		},
	}

	defenders = MSFTeam{
		Name:  "Defenders",
		Label: "DEF",
		Characters: []string{
			"luke-cage",
			"jessica-jones",
			"iron-fist",
			"daredevil",
			"punisher",
		},
	}

	defTron = MSFTeam{
		Name:  "Deftron",
		Label: "DEFTRON",
		Characters: []string{
			"ultron",
			"jessica-jones",
			"iron-fist",
			"daredevil",
			"punisher",
		},
	}

	fantasticFour = MSFTeam{
		Name:  "Fantastic Four",
		Label: "F4",
		Characters: []string{
			"invisible-woman",
			"the-thing",
			"human-torch",
			"mister-fantastic",
			"namor",
		},
	}

	guardians = MSFTeam{
		Name:  "Guardians of the Galaxy",
		Label: "GOG",
		Characters: []string{
			"star-lord",
			"groot",
			"drax",
			"rocket-racoon",
			"mantis",
		},
	}

	hydra = MSFTeam{
		Name:  "Hydra",
		Label: "HY",
		Characters: []string{
			"hydra-armored-guard",
			"hydra-scientist",
			"hydra-rifle-trooper",
			"hydra-sniper",
			"red-skull",
		},
	}

	hydraV2 = MSFTeam{
		Name:  "Hydra V2",
		Label: "HY2",
		Characters: []string{
			"hydra-grenadier",
			"winter-soldier",
			"kingpin",
			"crossbones",
		},
	}

	inhumans = MSFTeam{
		Name:  "Inhumans",
		Label: "INH",
		Characters: []string{
			"quake",
			"crystal",
			"karnak",
			"yoyo",
			"black-bolt",
		},
	}

	marauders = MSFTeam{
		Name:  "Marauders",
		Label: "MAR",
		Characters: []string{
			"mister-sinister",
			"sabretooth",
			"emmafrost",
			"mystique",
			"stryfe",
		},
	}

	mawTron = MSFTeam{
		Name:  "Mawtron",
		Label: "MAWTRON",
		Characters: []string{
			"ultron",
			"ebony-maw",
			"black-bolt",
			"minn-erva",
			"thanos",
		},
	}

	mercenaries = MSFTeam{
		Name:  "Mercenaries",
		Label: "MERC",
		Characters: []string{
			"taskmaster",
			"mercenary-sniper",
			"mercenary-riot-guard",
			"mercenary-lieutenant",
			"bullseye",
		},
	}

	powerArmorV2 = MSFTeam{
		Name:  "Power Armor V2",
		Label: "PA2",
		Characters: []string{
			"iron-man",
			"iron-heart",
			"war-machine",
			"falcon",
			"rescue",
		},
	}

	shieldCoulson = MSFTeam{
		Name:  "S.H.I.E.L.D. Coulson",
		Label: "SHC",
		Characters: []string{
			"nick-fury",
			"coulson",
			"shield-medic",
			"shield-security",
			"shield-assault",
		},
	}

	sinisterSix = MSFTeam{
		Name:  "Sinister 6",
		Label: "S6",
		Characters: []string{
			"rhino",
			"green-goblin",
			"shocker",
			"vulture",
			"mysterio",
		},
	}

	supernatural = MSFTeam{
		Name:  "Supernatural",
		Label: "SN",
		Characters: []string{
			"mordo",
			"ghost-rider",
			"scarlet-witch",
			"doctor-strange",
			"elsa-bloodstone",
		},
	}

	symbiotes = MSFTeam{
		Name:  "Symbiotes",
		Label: "SYM",
		Characters: []string{
			"spider-man-symbiote",
			"carnage",
			"venom",
			"spider-man",
			"spider-man-miles",
		},
	}

	symTech = MSFTeam{
		Name:  "SymTech",
		Label: "SYMTECH",
		Characters: []string{
			"spider-man-symbiote",
			"carnage",
			"venom",
			"scientist-supreme",
			"shuri",
		},
	}

	techWing = MSFTeam{
		Name:  "TechWing",
		Label: "TECHWING",
		Characters: []string{
			"ultron",
			"scientist-supreme",
			"shuri",
			"minn-erva",
			"falcon",
		},
	}

	ultronsAngels = MSFTeam{
		Name:  "Ultron's Angels",
		Label: "ANGELTRON",
		Characters: []string{
			"ultron",
			"shuri",
			"scientist-supreme",
			"invisible-woman",
			"minn-erva",
		},
	}

	wakandans = MSFTeam{
		Name:  "Wakandans",
		Label: "WAK",
		Characters: []string{
			"shuri",
			"black-panther",
			"killmonger",
			"mbaku",
			"okoye",
		},
	}

	xForce = MSFTeam{
		Name:  "X-Force",
		Label: "XFORCE",
		Characters: []string{
			"negasonic",
			"cable",
			"deadpool",
			"domino",
			"x23",
		},
	}

	xMen = MSFTeam{
		Name:  "X-Men",
		Label: "XMEN",
		Characters: []string{
			"wolverine",
			"colossus",
			"cyclops",
			"storm",
			"phoenix",
		},
	}
)

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}

	return config.Client(context.Background(), tok)
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)

	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}

	defer f.Close()

	json.NewEncoder(f).Encode(token)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)

	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}

	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}

	defer f.Close()

	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)

	return tok, err
}

func main() {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://msf.pal.gg/rest/v1/alliance/%s/characters", msfAllianceID), nil)
	if err != nil {
		log.Fatalln(err)
	}

	req.Header.Set("api-key", msfPalAPIKey)

	client := &http.Client{}
	playerCharactersResp, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}

	defer playerCharactersResp.Body.Close()

	data, err := ioutil.ReadAll(playerCharactersResp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	msfPlayerCharacters := MSFPlayerCharacters{}
	err = json.Unmarshal(data, &msfPlayerCharacters)
	if err != nil {
		log.Fatalln(err)
	}

	playerCharactersMap := map[string]MSFPlayerCharacters{}

	for _, character := range msfPlayerCharacters {
		if _, ok := playerCharactersMap[character.Player]; ok {
			playerCharactersMap[character.Player] = append(playerCharactersMap[character.Player], character)
		} else {
			playerCharactersMap[character.Player] = MSFPlayerCharacters{character}
		}
	}

	generateTopWarOffenseTeamsByPlayerReport(playerCharactersMap)
	generateTopWarDefenseTeamsByPlayerReport(playerCharactersMap)
	generateTopWarFlexTeamsByPlayerReport(playerCharactersMap)
	generateTopU7RaidTeamsByPlayerReport(playerCharactersMap)
}

func getSheetsService() *sheets.Service {
	if sheetsService != nil {
		return sheetsService
	}

	credentialsFile, err := ioutil.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(credentialsFile, "https://www.googleapis.com/auth/spreadsheets")
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}

	sheetsClient := getClient(config)

	sheetsService, err = sheets.New(sheetsClient)
	if err != nil {
		log.Fatalf("Unable to retrieve Sheets client: %v", err)
	}

	return sheetsService
}

func generateTopWarOffenseTeamsByPlayerReport(playerCharactersMap map[string]MSFPlayerCharacters) {
	teams := []MSFTeam{blackOrder, brotherhoodV2, defTron, fantasticFour, hydraV2, inhumans, powerArmorV2, supernatural, symbiotes, xForce, xMen}

	updateSheet(generateAverageTeamPowerByPlayerReport(playerCharactersMap, teams, "Offense"))
}

func generateTopWarDefenseTeamsByPlayerReport(playerCharactersMap map[string]MSFPlayerCharacters) {
	teams := []MSFTeam{asgardian, hydra, marauders, mercenaries, shieldCoulson, avengers, brawlers, sinisterSix}

	updateSheet(generateAverageTeamPowerByPlayerReport(playerCharactersMap, teams, "Defense"))
}

func generateTopWarFlexTeamsByPlayerReport(playerCharactersMap map[string]MSFPlayerCharacters) {
	teams := []MSFTeam{aim, defenders, guardians, wakandans}

	updateSheet(generateAverageTeamPowerByPlayerReport(playerCharactersMap, teams, "Flex"))
}

func generateTopU7RaidTeamsByPlayerReport(playerCharactersMap map[string]MSFPlayerCharacters) {
	teams := []MSFTeam{mawTron, symTech, techWing, ultronsAngels}

	updateSheet(generateAverageTeamPowerByPlayerReport(playerCharactersMap, teams, "U7"))
}

func generateAverageTeamPowerByPlayerReport(playerCharactersMap map[string]MSFPlayerCharacters, teams []MSFTeam, sheetName string) (writeRange string, valueRange *sheets.ValueRange) {
	playerTeamsMap := map[string][]int{}

	for _, team := range teams {
		for player, characters := range playerCharactersMap {
			if _, ok := playerTeamsMap[player]; !ok {
				playerTeamsMap[player] = []int{}
			}

			teamTotalPower := 0
			for _, teamCharacter := range team.Characters {
				for _, character := range characters {
					if strings.ToLower(character.ID) == teamCharacter {
						teamTotalPower += character.Power
					}
				}
			}

			playerTeamsMap[player] = append(playerTeamsMap[player], teamTotalPower)
		}
	}

	// Initialize cell values.
	cellValues := make([][]interface{}, len(playerTeamsMap)+1, len(playerTeamsMap)+1)
	for i := range cellValues {
		cellValues[i] = make([]interface{}, len(teams)+2, len(teams)+2)
	}

	// Setup table header row.
	cellValues[0] = []interface{}{"Player"}

	for _, team := range teams {
		cellValues[0] = append(cellValues[0], team.Label)
	}

	cellValues[0] = append(cellValues[0], "Average")

	// Sort player keys to ensure consistent ordering when iterating.
	playerKeys := make([]string, 0)
	for player := range playerTeamsMap {
		playerKeys = append(playerKeys, player)
	}

	sort.Strings(playerKeys)

	for columnIndex := 0; columnIndex < len(cellValues[0]); columnIndex++ {
		for playerIndex, player := range playerKeys {

			// First column contains names of players.
			if columnIndex == 0 {
				cellValues[playerIndex+1][columnIndex] = player
				continue
			}

			// Last column contains average total team power per player.
			if columnIndex == len(cellValues[0])-1 {
				totalPower := 0
				for _, teamTotalPower := range playerTeamsMap[player] {
					totalPower += teamTotalPower
				}

				cellValues[playerIndex+1][columnIndex] = totalPower / len(playerTeamsMap[player])
				continue
			}

			// Set the player's total power for a given team.
			cellValues[playerIndex+1][columnIndex] = playerTeamsMap[player][columnIndex-1]
		}
	}

	sort.Sort(ByAverageTotalTeamPower(cellValues))

	valueRange = &sheets.ValueRange{
		MajorDimension: "ROWS",
		Values:         cellValues,
	}

	writeRange = fmt.Sprintf("%s!A1:%s%s", sheetName, sheetCellIndexToLetter[len(cellValues[0])-1], strconv.Itoa(len(cellValues)))

	return
}

func updateSheet(writeRange string, valueRange *sheets.ValueRange) {
	srv := getSheetsService()

	_, err := srv.Spreadsheets.Values.Update(spreadsheetID, writeRange, valueRange).ValueInputOption("RAW").Do()
	if err != nil {
		log.Fatalf("Unable to update data on sheet: %v", err)
	}
}

func prettyPrintJSON(data interface{}) {
	formattedData, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		log.Fatalln(err)
	}

	log.Println(string(formattedData))
}
