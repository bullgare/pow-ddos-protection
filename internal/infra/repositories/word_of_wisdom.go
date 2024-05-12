package repositories

import (
	"math/rand"

	"github.com/bullgare/pow-ddos-protection/internal/domain/contracts"
)

func NewWOW() WOW {
	// taken from https://www.churchofjesuschrist.org/bc/content/ldsorg/content/english/manual/missionary/pdf/36953_the-word-of-wisdom-eng.pdf
	return WOW{
		quotes: []string{
			"A Word of Wisdom, for the benefit of the . . . church, and also the saints in Zion . . .",
			"Given for a principle with promise, adapted to the capacity of the weak and the weakest of all saints. . . .",
			"Behold, verily, thus saith the Lord unto you: In consequence of evils and designs which do and will exist in the hearts of conspiring men in the last days, I . . . forewarn you, by giving unto you this word of wisdom by revelation",
			"That inasmuch as any man drinketh wine or strong drink among you, behold it is not good. . . .",
			"And again, tobacco is not for the body, neither for the belly, and is not good for man. . . .",
			"And again, hot drinks [tea and coffee] are not for the body or belly.",
			". . . All wholesome herbs God hath ordained for the constitution, nature, and use of man. . . .",
			"Yea, flesh also of beasts and of the fowls of the air, I, the Lord, have ordained for the use of man with thanksgiving; nevertheless they are to be used sparingly. . . .",
			"All grain is good for the food of man; as also the fruit of the vine; that which yieldeth fruit, whether in the ground or above the ground. . . .",
			"And all saints who remember to keep and do these sayings . . . shall receive health in their navel and marrow to their bones;",
			"And shall find wisdom and great treasures of knowledge, even hidden treasures;",
			"And shall run and not be weary, and shall walk and not faint.",
			"And I, the Lord, give unto them a promise, that the destroying angel shall pass by them, as the children of Israel, and not slay them. Amen",
		},
	}
}

var _ contracts.WOWQuotes = WOW{}

type WOW struct {
	quotes []string
}

func (w WOW) GetRandomQuote() string {
	idx := rand.Intn(len(w.quotes))
	return w.quotes[idx]
}
