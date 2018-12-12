# fortnite-go
An interfacer for the [Fortnite](https://www.epicgames.com/fortnite) game API. Can retrieve player statistics and global leaderboard information by specified platform type. Handles authentication, token renewal, and token destruction upon program exit.

🍻 [Tip me!](https://www.paypal.me/wh93?country.x=US&locale.x=en_US)

## Setup
In order to authenticate successfully with the Epic game servers, you must have an Epic Account, and Fortnite installed on the PC you are using. Two authentication tokens must be extracted on client and game launch.

To obtain header tokens:
*   Install & Open [Fiddler 4](https://www.telerik.com/download/fiddler)
*   In Tools -> Options -> HTTPS, enable "Capture HTTPS CONNECTs" and "Decrypt HTTPS traffic"
*   Start your Epic Games Launcher.
*   You will see a request with _/account/api/oauth/token_. Click on it and then click `Inspectors` tab to get the header (Copy `Authorization` header content and remove "basic ") => **This header is your Launcher Token**
*   Launch Fortnite
*   You will see again a request with _/account/api/oauth/token_. Click on it and then click `Inspectors` tab to get the header (Copy `Authorization` header content and remove "basic ") => **This header is your Game Token**

## Usage

See [Godoc](https://godoc.org/github.com/henhouse/fortnite-go) for in-depth documentation.

### Player Stats
To retrieve a player's information and statistics for Battle Royale:
```go
// Create the session.
sess := fornitego.Create("USERNAME", "PASSWORD", "LAUNCHER-TOKEN", "GAME-TOKEN")

// Retrieve player info and stats by Username and Platform.
player, err := s.QueryPlayer("PlayerName", "", fornitego.PC) // (PC/Xbox/PS4)
if err != nil {
	fmt.Println(err)
}

// Retrieve player info and stats by Account ID and Platform.
player, err := s.QueryPlayer("", "AccountID", fornitego.PC) // (PC/Xbox/PS4)
if err != nil {
	fmt.Println(err)
}
```
If the player exists, a result may look like the example below. (Represented in JSON)
```json
{
  "AccountInfo": {
    "AccountID": "6cd40c1722f2497fa1d2145b26da88e3",
    "Username": "WalterJr2",
    "Platform": "pc"
  },
  "Stats": {
    "Solo": {
      "Wins": 23,
      "Top10": 86,
      "Top25": 154,
      "KillDeathRatio": "3.13",
      "WinPercentage": "6.74",
      "Matches": 341,
      "Kills": 995,
      "MinutesPlayed": 2174,
      "KillsPerMatch": "2.92",
      "KillsPerMinute": "0.46",
      "Score": 56247
    },
    "Duo": {
      "Wins": 45,
      "Top5": 89,
      "Top12": 149,
      "KillDeathRatio": "3.27",
      "WinPercentage": "11.03",
      "Matches": 408,
      "Kills": 1186,
      "MinutesPlayed": 1465,
      "KillsPerMatch": "2.91",
      "KillsPerMinute": "0.81",
      "Score": 91499
    },
    "Squad": {
      "Wins": 116,
      "Top3": 190,
      "Top6": 305,
      "KillDeathRatio": "3.60",
      "WinPercentage": "14.23",
      "Matches": 815,
      "Kills": 2516,
      "MinutesPlayed": 3143,
      "KillsPerMatch": "3.09",
      "KillsPerMinute": "0.80",
      "Score": 253462
    }
  }
}
```

### Leaderboard
To retrieve the top 50 global wins leaderboard:
```go
lb, err := sess.GetWinsLeaderboard(fornitego.PC, fornitego.Squad) // (Solo, Duo, Squad)
if err != nil {
	fmt.Println(err)
}
```

A typical response would look like:
```json
[
  {
    "DisplayName": "4hs_UwatakashiТV",
    "Rank": 1,
    "Wins": 1131
  },
  {
    "DisplayName": "qoowill",
    "Rank": 2,
    "Wins": 827
  },
  {
    "DisplayName": "RedemeЯ",
    "Rank": 3,
    "Wins": 818
  },
  {
    "DisplayName": "Copy - TH",
    "Rank": 4,
    "Wins": 801
  },
  {
    "DisplayName": "TTV.vannesskwan",
    "Rank": 5,
    "Wins": 800
  },
  {
    "DisplayName": "BlooTeaTV",
    "Rank": 6,
    "Wins": 789
  },
  {
    "DisplayName": "Twitch_PuZiiyo",
    "Rank": 7,
    "Wins": 765
  },
  {
    "DisplayName": "Infamous Uniq",
    "Rank": 8,
    "Wins": 680
  },
  {
    "DisplayName": "ŁїƒεSnoopySworld",
    "Rank": 9,
    "Wins": 635
  },
  {
    "DisplayName": "tuổi lz sánh vai",
    "Rank": 10,
    "Wins": 622
  },
  {
    "DisplayName": "SaltySoji",
    "Rank": 11,
    "Wins": 620
  },
  {
    "DisplayName": "Fluuuuuuu",
    "Rank": 12,
    "Wins": 619
  },
  {
    "DisplayName": "Faze_MadGames",
    "Rank": 13,
    "Wins": 618
  },
  {
    "DisplayName": "Mafia WillzonePH",
    "Rank": 14,
    "Wins": 618
  },
  {
    "DisplayName": "IDOLˆMrTìnhˆ",
    "Rank": 15,
    "Wins": 609
  },
  {
    "DisplayName": "Twitch.DapanoTV",
    "Rank": 16,
    "Wins": 607
  },
  {
    "DisplayName": "VIP. Trung Lương",
    "Rank": 17,
    "Wins": 599
  },
  {
    "DisplayName": "Ĺαšt-Prince",
    "Rank": 18,
    "Wins": 576
  },
  {
    "DisplayName": "DongminHero_o",
    "Rank": 19,
    "Wins": 565
  },
  {
    "DisplayName": "TacoSlut.",
    "Rank": 20,
    "Wins": 563
  },
  {
    "DisplayName": "Ĺαšt-ÐεÑz ツ",
    "Rank": 21,
    "Wins": 554
  },
  {
    "DisplayName": "Pvt.alifrizani",
    "Rank": 22,
    "Wins": 544
  },
  {
    "DisplayName": "Copy - 2 TAP",
    "Rank": 23,
    "Wins": 543
  },
  {
    "DisplayName": "Death Donator",
    "Rank": 24,
    "Wins": 538
  },
  {
    "DisplayName": "Pvt.Brokutt",
    "Rank": 25,
    "Wins": 534
  },
  {
    "DisplayName": "ⓛⓞⓥⓔBetty-CosMix",
    "Rank": 26,
    "Wins": 533
  },
  {
    "DisplayName": "Pre3idium",
    "Rank": 27,
    "Wins": 528
  },
  {
    "DisplayName": "BC Mr.Spawnz",
    "Rank": 28,
    "Wins": 528
  },
  {
    "DisplayName": "CTGS.TTâm SoNy",
    "Rank": 29,
    "Wins": 525
  },
  {
    "DisplayName": "Early_Morning",
    "Rank": 30,
    "Wins": 518
  },
  {
    "DisplayName": "Kreyzi",
    "Rank": 31,
    "Wins": 516
  },
  {
    "DisplayName": "i7 - 1080 Ti",
    "Rank": 32,
    "Wins": 510
  },
  {
    "DisplayName": "sunin-",
    "Rank": 33,
    "Wins": 509
  },
  {
    "DisplayName": "1.0.1",
    "Rank": 34,
    "Wins": 502
  },
  {
    "DisplayName": "Sarkanos",
    "Rank": 35,
    "Wins": 492
  },
  {
    "DisplayName": "VexNguyen",
    "Rank": 36,
    "Wins": 487
  },
  {
    "DisplayName": "eShield DoNtm1nd",
    "Rank": 37,
    "Wins": 481
  },
  {
    "DisplayName": "Ninja O Ceifador",
    "Rank": 38,
    "Wins": 480
  },
  {
    "DisplayName": "Fa_YeuVoBan",
    "Rank": 39,
    "Wins": 473
  },
  {
    "DisplayName": "N68 3 .",
    "Rank": 40,
    "Wins": 472
  },
  {
    "DisplayName": "K- Mày Tuổi Tôm",
    "Rank": 41,
    "Wins": 471
  },
  {
    "DisplayName": "Oraculoo",
    "Rank": 42,
    "Wins": 471
  },
  {
    "DisplayName": "i7 - 6700K",
    "Rank": 43,
    "Wins": 466
  },
  {
    "DisplayName": "Fable Zamas",
    "Rank": 44,
    "Wins": 464
  },
  {
    "DisplayName": "NoCry- arT.",
    "Rank": 45,
    "Wins": 461
  },
  {
    "DisplayName": "TwitchTranq96",
    "Rank": 46,
    "Wins": 457
  },
  {
    "DisplayName": "Twitch.ItsWiKKiD",
    "Rank": 47,
    "Wins": 455
  },
  {
    "DisplayName": "이영돈",
    "Rank": 48,
    "Wins": 455
  },
  {
    "DisplayName": "Mafia Chrism-",
    "Rank": 49,
    "Wins": 455
  },
  {
    "DisplayName": "Infamous G0D",
    "Rank": 50,
    "Wins": 454
  }
]
```

---
### Special Thanks
To [qlaffont](https://github.com/qlaffont) for [fortnite-api](https://github.com/qlaffont/fortnite-api), which this project was largely based off of and inspired by.
