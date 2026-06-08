// Mule pet script
// The mule is a sturdy, stubborn pack animal that occasionally makes its feelings known.

/**
 * Called approximately once every 10 rounds while the pet is active.
 * @param {PetObject} pet - The pet.
 * @param {ActorObject} actor - The pet's owner.
 * @param {RoomObject} room - The room the pet is in.
 * @returns {void}
 */
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
