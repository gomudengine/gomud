// Cat pet script
// The cat is an aloof companion that occasionally deigns to acknowledge its owner.

// PetAct is called approximately once every 10 rounds.
function PetAct(pet, actor, room) {
    var actions = [
        pet.NameSimple() + ' grooms itself with practiced indifference.',
        pet.NameSimple() + ' stares at something only it can see.',
        pet.NameSimple() + ' flicks its tail once, slowly.',
        pet.NameSimple() + ' stretches languidly and yawns.',
        pet.NameSimple() + ' blinks at ' + actor.GetCharacterName(false) + ' with half-closed eyes.',
    ];
    room.SendText(actions[RandInt(0, actions.length - 1)]);
}
