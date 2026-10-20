package names

import (
	"fmt"
	"math/rand"
	"time"
)

var left = []string{
	"admiring", "adoring", "affectionate", "agitated", "amazing",
	"angry", "awesome", "beautiful", "blissful", "bold",
	"brave", "busy", "charming", "clever", "cool",
	"compassionate", "competent", "confident", "cranky", "crazy",
	"dazzling", "determined", "distracted", "dreamy", "eager",
	"ecstatic", "elastic", "elated", "elegant", "eloquent",
	"epic", "exciting", "fervent", "festive", "flamboyant",
	"focused", "friendly", "frosty", "funny", "gallant",
	"gifted", "goofy", "gracious", "great", "happy",
	"hardcore", "heuristic", "hopeful", "hungry", "infallible",
	"inspiring", "intelligent", "interesting", "jolly", "jovial",
	"keen", "kind", "laughing", "loving", "lucid",
	"magical", "mystifying", "modest", "musing", "naughty",
	"nervous", "nice", "nifty", "nostalgic", "objective",
	"optimistic", "peaceful", "pedantic", "pensive", "practical",
	"priceless", "quirky", "quizzical", "recursing", "relaxed",
	"reverent", "romantic", "sad", "serene", "sharp",
	"silly", "sleepy", "stoic", "suspicious", "sweet",
	"tender", "thirsty", "trusting", "unruffled", "upbeat",
	"vibrant", "vigilant", "vigorous", "wizardly", "wonderful",
	"zealous", "zen",
}

var right = []string{
	"albattani", "allen", "almeida", "antonelli", "archimedes",
	"ardinghelli", "aryabhata", "austin", "babbage", "banach",
	"bardeen", "bartik", "bassi", "beaver", "bell",
	"benz", "bhabha", "bhaskara", "black", "blackburn",
	"blackwell", "bohr", "booth", "borg", "bose",
	"bouman", "bowman", "boyd", "brahmagupta", "brattain",
	"brown", "burnell", "cannon", "carson", "cartwright",
	"chandrasekhar", "cerf", "chatelet", "chatterjee", "chebyshev",
	"cohen", "colden", "cori", "cray", "curie",
	"darwin", "davinci", "dewdney", "dhawan", "diffie",
	"dijkstra", "dirac", "driscoll", "dubinsky", "easley",
	"edison", "einstein", "elbakyan", "elgamal", "elion",
	"ellis", "engelbart", "euclid", "euler", "faraday",
	"feistel", "fermat", "fermi", "feynman", "franklin",
	"gagarin", "galileo", "galois", "ganguly", "gates",
	"goldberg", "goldstine", "goldwasser", "golick", "goodall",
	"gould", "greider", "grothendieck", "haibt", "hamilton",
	"haslett", "hawking", "hellman", "heisenberg", "hermann",
	"herschel", "hertz", "heyrovsky", "hodgkin", "hofstadter",
	"hoover", "hopper", "hugle", "hypatia", "ishizaka",
	"jackson", "jang", "jemison", "jennings", "jepsen",
	"johnson", "joliot", "jones", "kalam", "kapitsa",
	"kare", "keldysh", "keller", "kepler", "khayyam",
	"khorana", "kilby", "knuth", "kowalevski", "lalande",
	"lamarr", "lamport", "leakey", "leavitt", "lederberg",
	"lehmann", "lewin", "lichterman", "liskov", "lovelace",
	"lumiere", "mahavira", "margulis", "matsumoto", "maxwell",
	"mayer", "mccarthy", "mcclintock", "mclaren", "mclean",
	"mcnulty", "mendel", "mendeleev", "meitner", "meninsky",
	"merkle", "mestorf", "mirzakhani", "moore", "morse",
	"murdock", "moser", "napier", "nash", "neumann",
	"newton", "nightingale", "nobel", "noether", "northcutt",
	"noyce", "panini", "pare", "pascal", "pasteur",
	"payne", "perlman", "pike", "poincare", "poitras",
	"proskuriakova", "ptolemy", "raman", "ramanujan", "ride",
	"ritchie", "rhodes", "robinson", "roentgen", "rosalind",
	"rubin", "saha", "sammet", "sanderson", "satoshi",
	"shamir", "shannon", "shaw", "shirley", "shockley",
	"shternberg", "sinoussi", "snyder", "solomon", "spence",
	"stallman", "stonebraker", "sutherland", "swanson", "swartz",
	"swirles", "taussig", "tereshkova", "tesla", "tharp",
	"thompson", "torvalds", "tu", "turing", "tsang",
	"varahamihira", "vaughan", "villani", "visvesvaraya", "volhard",
	"wescoff", "wilbur", "wiles", "williams", "williamson",
	"wilson", "wing", "wozniak", "wright", "wu",
	"yalow", "yonath", "zhukovsky",
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

// GenerateRandomName generates a random container name like "clever_curie"
func GenerateRandomName() string {
	l := left[rand.Intn(len(left))]
	r := right[rand.Intn(len(right))]
	return fmt.Sprintf("%s_%s", l, r)
}
