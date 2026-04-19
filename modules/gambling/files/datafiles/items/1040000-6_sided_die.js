
const FACE_NAMES = ["one", "two", "three", "four", "five", "six"];

const ROLL_MESSAGES = [
    "{name} gives their {item} a flick and it skitters across the floor.",
    "{name} rattles their {item} in a cupped hand and lets it fly.",
    "{name} tosses their {item} with a practiced spin.",
    "{name} blows on their {item} for luck and sends it rolling.",
];

function onCommand_roll(user, item, room) {

    var result = UtilDiceRoll(1, 6);
    var faceName = FACE_NAMES[result - 1];

    var msgIdx = UtilDiceRoll(1, ROLL_MESSAGES.length) - 1;
    var flavorMsg = ROLL_MESSAGES[msgIdx]
        .replace("{name}", user.GetCharacterName(true))
        .replace("{item}", item.Name(true));

    SendUserMessage(
        user.UserId(),
        "You roll your <ansi fg=\"item\">" + item.Name(true) + "</ansi> and it lands on <ansi fg=\"yellow-bold\">" + faceName + " (" + result + ")</ansi>."
    );

    SendRoomMessage(
        room.RoomId(),
        flavorMsg + " It lands on <ansi fg=\"yellow-bold\">" + faceName + " (" + result + ")</ansi>.",
        user.UserId()
    );

    return true;
}
