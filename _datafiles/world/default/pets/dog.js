// Dog pet script
// The dog is a loyal companion that reacts to its owner's actions.

// PetAct is called approximately once every 10 rounds.
// No top-level probability check is needed here; add one if you want
// behaviour that fires less frequently than that.
function PetAct(pet, actor, room) {
    var actions = [
        pet.NameSimple() + ' wags its tail happily.',
        pet.NameSimple() + ' sniffs around the room.',
        pet.NameSimple() + ' sits at ' + actor.GetCharacterName(false) + "'s feet.",
        pet.NameSimple() + ' lets out a soft woof.',
        pet.NameSimple() + ' nudges ' + actor.GetCharacterName(false) + "'s hand.",
    ];
    room.SendText(actions[RandInt(0, actions.length - 1)]);
}
