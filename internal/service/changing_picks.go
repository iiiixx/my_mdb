package service

type yearRangePick struct {
	kind  string
	title string
	from  int
	to    int
}

var yearRangePicks = []yearRangePick{
	{kind: "seventies", title: "Фильмы 70-ых", from: 1970, to: 1979},
	{kind: "eighties", title: "Фильмы 80-ых", from: 1980, to: 1989},
	{kind: "nineties", title: "Фильмы 90-ых", from: 1990, to: 1999},
	{kind: "zeros", title: "Фильмы 2000-ых", from: 2000, to: 2009},
	{kind: "tens", title: "Фильмы 2010-ых", from: 2010, to: 2016},
	{kind: "classic_50_60", title: "Классика 50–60-х", from: 1950, to: 1969},
	{kind: "turn_of_millennium", title: "На рубеже тысячелетий", from: 1998, to: 2004},
}

type tagPick struct {
	kind  string
	title string
	query string
}

type tagCategory struct {
	kind  string
	title string
	picks []tagPick
}

var tagCategories = []tagCategory{
	{
		kind:  "crime",
		title: "Криминальное",
		picks: []tagPick{
			{kind: "detective", title: "Детективы", query: "detective"},
			{kind: "serial_killer", title: "Про серийных убийц", query: "serial killer"},
			{kind: "mafia", title: "Мафия и гангстеры", query: "mafia"},
			{kind: "heist", title: "Ограбления", query: "heist"},
			{kind: "prison", title: "Фильмы про тюрьму", query: "prison"},
			{kind: "courtroom", title: "Суд и адвокаты", query: "courtroom"},
		},
	},
	{
		kind:  "thriller",
		title: "Триллеры",
		picks: []tagPick{
			{kind: "tense", title: "Держит в напряжении", query: "tense"},
			{kind: "plot_twist", title: "Неожиданная развязка", query: "plot twist"},
			{kind: "psychological", title: "Психологическое", query: "psychological"},
			{kind: "suspense", title: "Саспенс", query: "suspense"},
			{kind: "dark", title: "Мрачное", query: "dark"},
		},
	},
	{
		kind:  "scifi_fantasy",
		title: "Фантастика и миры",
		picks: []tagPick{
			{kind: "space", title: "Космос", query: "space"},
			{kind: "aliens", title: "Инопланетяне", query: "aliens"},
			{kind: "time_travel", title: "Путешествия во времени", query: "time travel"},
			{kind: "dystopia", title: "Антиутопия", query: "dystopia"},
			{kind: "post_apocalyptic", title: "Постапокалипсис", query: "post apocalyptic"},
			{kind: "cyberpunk", title: "Киберпанк", query: "cyberpunk"},
			{kind: "fantasy_world", title: "Фэнтези-мир", query: "fantasy world"},
		},
	},
	{
		kind:  "feelings",
		title: "Настроение",
		picks: []tagPick{
			{kind: "humor", title: "С юмором", query: "humor"},
			{kind: "black_comedy", title: "Чёрная комедия", query: "black comedy"},
			{kind: "satire", title: "Сатира", query: "satire"},
			{kind: "romantic", title: "Романтика", query: "romantic"},
			{kind: "self_discovery", title: "Самопознание", query: "self discovery"},
			{kind: "coming_of_age", title: "Взросление", query: "coming of age"},
		},
	},
	{
		kind:  "history_war",
		title: "Историческое",
		picks: []tagPick{
			{kind: "world_war_ii", title: "Вторая мировая", query: "world war ii"},
			{kind: "war", title: "Про войну", query: "war"},
			{kind: "based_on_true_story", title: "Основано на реальных событиях", query: "based on true story"},
			{kind: "biography", title: "Биографии", query: "biography"},
			{kind: "politics", title: "Политика", query: "politics"},
		},
	},
	{
		kind:  "style",
		title: "Визуал и стиль",
		picks: []tagPick{
			{kind: "visually_stunning", title: "Визуально впечатляющие", query: "visually stunning"},
			{kind: "cgi", title: "CGI", query: "cgi"},
			{kind: "stylized", title: "Стилизация", query: "stylized"},
			{kind: "cinematography", title: "Кинематографично", query: "cinematography"},
			{kind: "atmospheric", title: "Атмосферное", query: "atmospheric"},
		},
	},
}
