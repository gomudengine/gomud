// Owl pet script
// The owl is a wise, observant companion that notices things others miss.

// PetAct is called approximately once every 10 rounds.
function PetAct(pet, actor, room) {
    var actions = [
        pet.NameSimple() + ' rotates its head and surveys the room with amber eyes.',
        pet.NameSimple() + ' ruffles its feathers and settles back into stillness.',
        pet.NameSimple() + ' emits a soft, low hoot.',
        pet.NameSimple() + ' bobs its head, seemingly deep in thought.',
        pet.NameSimple() + ' clicks its beak once.',
    ];
    room.SendText(actions[RandInt(0, actions.length - 1)]);
}
