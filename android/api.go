package android

import (
	"strconv"

	"github.com/bwmarrin/discordgo"
)

var levels = map[int]string{
	1:  "Android 1.0",
	2:  "Android 1.1: https://developer.android.com/about/versions/android-1.1.html",
	3:  "Android 1.5 Cupcake: https://developer.android.com/about/versions/android-1.5.html",
	4:  "Android 1.6 Donut: https://developer.android.com/about/versions/android-1.6.html",
	5:  "Android 2.0 Eclair: https://developer.android.com/about/versions/android-2.0.html",
	6:  "Android 2.0.1 Eclair: https://developer.android.com/about/versions/android-2.0.1.html",
	7:  "Android 2.1 Eclair: https://developer.android.com/about/versions/android-2.1.html",
	8:  "Android 2.2 Froyo: https://developer.android.com/about/versions/android-2.2.html",
	9:  "Android 2.3 Gingerbread: https://developer.android.com/about/versions/android-2.3.html",
	10: "Android 2.3.3 Gingerbread: https://developer.android.com/about/versions/android-2.3.3.html",
	11: "Android 3.0 Honeycomb: https://developer.android.com/about/versions/android-3.0.html",
	12: "Android 3.1 Honeycomb: https://developer.android.com/about/versions/android-3.1.html",
	13: "Android 3.2 Honeycomb: https://developer.android.com/about/versions/android-3.2.html",
	14: "Android 4.0 Ice Cream Sandwich: https://developer.android.com/about/versions/android-4.0.html",
	15: "Android 4.0.3 Ice Cream Sandwich: https://developer.android.com/about/versions/android-4.0.3.html",
	16: "Android 4.1 Jelly Bean: https://developer.android.com/about/versions/android-4.1.html",
	17: "Android 4.2 Jelly Bean: https://developer.android.com/about/versions/android-4.2.html",
	18: "Android 4.3 Jelly Bean: https://developer.android.com/about/versions/android-4.3.html",
	19: "Android 4.4 KitKat: https://developer.android.com/about/versions/android-4.4.html",
	20: "Android 4.4W KitKat",
	21: "Android 5.0 Lollipop: https://developer.android.com/about/versions/android-5.0.html",
	22: "Android 5.1 Lollipop: https://developer.android.com/about/versions/android-5.1.html",
	23: "Android 6.0 Marshmallow: https://developer.android.com/about/versions/marshmallow/android-6.0.html",
	24: "Android 7.0 Nougat: https://developer.android.com/about/versions/nougat/android-7.0.html",
	25: "Android 7.1 Nougat: https://developer.android.com/about/versions/nougat/android-7.1.html",
	26: "Android 8.0 Oreo: https://developer.android.com/about/versions/oreo/android-8.0.html",
	27: "Android 8.1 Oreo: https://developer.android.com/about/versions/oreo/android-8.1.html",
	28: "Android 9.0 Pie: https://developer.android.com/about/versions/pie/android-9.0",
}

func APILevel(args []string, s *discordgo.Session, m *discordgo.Message) {
	level, err := strconv.Atoi(args[0])

	text := levels[level]

	if err != nil || text == "" {
		text = "Unknown"
	}

	s.ChannelMessageSend(m.ChannelID, text)
}
