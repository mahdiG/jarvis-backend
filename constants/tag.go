package constants

import "jarvis/models"

var TagJournal = models.Tag{
	Base: models.Base{
		ID: "journal",
	},
	Name: "journal",
}

var TagMusic = models.Tag{
	Base: models.Base{
		ID: "music",
	},
	Name: "music",
}

var TagGuitar = models.Tag{
	Base: models.Base{
		ID: "guitar",
	},
	Name: "guitar",
}

var TagPiano = models.Tag{
	Base: models.Base{
		ID: "piano",
	},
	Name: "piano",
}

var TagSing = models.Tag{
	Base: models.Base{
		ID: "sing",
	},
	Name: "sing",
}

var TagVoice = models.Tag{
	Base: models.Base{
		ID: "voice",
	},
	Name: "voice",
}
