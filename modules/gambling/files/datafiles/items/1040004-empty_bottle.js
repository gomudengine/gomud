
var SPIN_INTROS = [
    "You set the <ansi fg=\"item\">{item}</ansi> on its side and give it a sharp flick.",
    "You place the <ansi fg=\"item\">{item}</ansi> flat and spin it with two fingers.",
    "You set the <ansi fg=\"item\">{item}</ansi> down and send it spinning with a practiced twist.",
    "You give the <ansi fg=\"item\">{item}</ansi> a slow, deliberate spin.",
    "You flick the <ansi fg=\"item\">{item}</ansi> into a lazy spin.",
];

var SPIN_SETTLES = [
    "It wobbles, slows, and finally comes to rest pointing at",
    "It spins for a long moment before settling on",
    "It makes three full rotations and stops, pointing squarely at",
    "It slows, teeters, and tips toward",
    "It spins with surprising speed before coming to an abrupt stop, aimed at",
    "It drifts to a halt, the mouth of the bottle aimed directly at",
];

var SPIN_EMPTY = [
    "It spins freely and comes to rest pointing at nothing in particular. The room holds its breath, then lets it out.",
    "It slows and stops aimed at an empty stretch of floor. Somehow, this feels significant.",
    "It spins, slows, and points at absolutely no one. A moment of profound anticlimax.",
];

function onCommand_spin(user, item, room) {

    var players = room.GetPlayers();
    var mobs    = room.GetMobs();
    var targets = players.concat(mobs);

    var introMsg = SPIN_INTROS[UtilDiceRoll(1, SPIN_INTROS.length) - 1]
        .replace("{item}", item.Name(true));

    if (targets.length === 0) {
        var emptyMsg = SPIN_EMPTY[UtilDiceRoll(1, SPIN_EMPTY.length) - 1];
        SendUserMessage(user.UserId(), introMsg + "\n" + emptyMsg);
        SendRoomMessage(
            room.RoomId(),
            user.GetCharacterName(true) + " spins their <ansi fg=\"item\">" + item.Name(true) + "</ansi>. It points at no one.",
            user.UserId()
        );
        return true;
    }

    var target  = targets[UtilDiceRoll(1, targets.length) - 1];
    var settles = SPIN_SETTLES[UtilDiceRoll(1, SPIN_SETTLES.length) - 1];

    SendUserMessage(
        user.UserId(),
        introMsg + "\n" + settles + " <ansi fg=\"username\">" + target.GetCharacterName(false) + "</ansi>."
    );

    SendRoomMessage(
        room.RoomId(),
        user.GetCharacterName(true) + " spins their <ansi fg=\"item\">" + item.Name(true) + "</ansi>. " +
        settles + " " + target.GetCharacterName(true) + ".",
        user.UserId()
    );

    return true;
}
