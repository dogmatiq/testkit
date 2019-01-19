package fact

// Observer is a function is called when a fact is recorded.
type Observer func(Fact)

// ObserverSet is a collection of observers that can be notified as a group.
type ObserverSet []Observer

// Notify notifies all of the observers in the set of each of the given facts.
func (s ObserverSet) Notify(facts ...Fact) {
	for _, f := range facts {
		for _, o := range s {
			o(f)
		}
	}
}
