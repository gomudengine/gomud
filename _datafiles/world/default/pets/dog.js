// Dog pet script
// The dog is a loyal companion that reacts to its owner's actions.

/**
 * Called approximately once every 10 rounds while the pet is active.
 * @param {PetObject} pet - The pet.
 * @param {ActorObject} actor - The pet's owner.
 * @param {RoomObject} room - The room the pet is in.
 * @returns {void}
 */
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
