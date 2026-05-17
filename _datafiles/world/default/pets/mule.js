// Mule pet script
// The mule is a sturdy, stubborn pack animal that occasionally makes its feelings known.

// PetAct is called approximately once every 10 rounds.
function PetAct(pet, actor, room) {
    var actions = [
        pet.NameSimple() + ' stamps a hoof impatiently.',
        pet.NameSimple() + ' snorts and shakes its head.',
        pet.NameSimple() + ' shifts its load and huffs.',
        pet.NameSimple() + ' glances at ' + actor.GetCharacterName(false) + ' with a long-suffering expression.',
        pet.NameSimple() + ' swishes its tail at a fly.',
    ];
    room.SendText(actions[RandInt(0, actions.length - 1)]);
}
