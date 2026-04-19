
function onCommand_flip(user, item, room) {

    var skulls = UtilDiceRoll(1, 2) === 1;
    var result = skulls ? "skull" : "ship";

    SendUserMessage(
        user.UserId(),
        "You flip your <ansi fg=\"item\">" + item.Name(true) + "</ansi> into the air... it lands on <ansi fg=\"yellow-bold\">" + result + "</ansi>."
    );

    SendRoomMessage(
        room.RoomId(),
        user.GetCharacterName(true) + " flips their <ansi fg=\"item\">" + item.Name(true) + "</ansi>. It lands on <ansi fg=\"yellow-bold\">" + result + "</ansi>.",
        user.UserId()
    );

    return true;
}

// Each entry has a "you" version (seen by the player) and a "they" version (seen by the room).
var ROLL_FLAVOR = [
    {
        you:  "with the practiced ease of someone who has done this a thousand times",
        they: "with the practiced ease of someone who has done this a thousand times",
    },
    {
        you:  "absently, as if your fingers have a mind of their own",
        they: "absently, as if their fingers have a mind of their own",
    },
    {
        you:  "with a soft, rhythmic click against each knuckle",
        they: "with a soft, rhythmic click against each knuckle",
    },
    {
        you:  "slowly, letting the skull and ship catch the light in turn",
        they: "slowly, letting the skull and ship catch the light in turn",
    },
    {
        you:  "without even glancing down, your eyes fixed somewhere far away",
        they: "without even glancing down, their eyes fixed somewhere far away",
    },
    {
        you:  "back and forth, back and forth, the motion as easy as breathing",
        they: "back and forth, back and forth, the motion as easy as breathing",
    },
    {
        you:  "with a flick that sends it dancing from index to pinky and back again",
        they: "with a flick that sends it dancing from index to pinky and back again",
    },
    {
        you:  "so smoothly it barely seems to touch your skin at all",
        they: "so smoothly it barely seems to touch their skin at all",
    },
    {
        you:  "while humming something tuneless under your breath",
        they: "while humming something tuneless under their breath",
    },
    {
        you:  "your knuckles dipping and rising like a tiny wave carrying the coin along",
        they: "their knuckles dipping and rising like a tiny wave carrying the coin along",
    },
    {
        you:  "with a contemplative look, as if the coin is helping you think something through",
        they: "with a contemplative look, as if the coin is helping them think something through",
    },
    {
        you:  "idly, the skull winking in and out of view with each pass",
        they: "idly, the skull winking in and out of view with each pass",
    },
    {
        you:  "with surprising speed, the coin a golden blur across your hand",
        they: "with surprising speed, the coin a golden blur across their hand",
    },
    {
        you:  "letting it pause dramatically on your middle knuckle before continuing",
        they: "letting it pause dramatically on their middle knuckle before continuing",
    },
    {
        you:  "as if the weight of it is the only thing keeping you grounded right now",
        they: "as if the weight of it is the only thing keeping them grounded right now",
    },
    {
        you:  "with the unhurried calm of someone who has nowhere better to be",
        they: "with the unhurried calm of someone who has nowhere better to be",
    },
    {
        you:  "your expression unreadable, the coin moving like a secret between your fingers",
        they: "their expression unreadable, the coin moving like a secret between their fingers",
    },
    {
        you:  "so casually it seems almost insulting how good you are at it",
        they: "so casually it seems almost insulting how good they are at it",
    },
    {
        you:  "with a faint smile, like the coin just reminded you of something",
        they: "with a faint smile, like the coin just reminded them of something",
    },
];

var DROP_FLAVOR = [
    {
        you:  "but it slips between your fingers and <ansi fg=\"yellow\">clatters to the ground</ansi>",
        they: "then fumbles it. The coin <ansi fg=\"yellow\">spins to the floor</ansi> with a ringing clatter",
    },
    {
        you:  "until it catches the edge of your knuckle wrong and <ansi fg=\"yellow\">bounces free</ansi>, skittering across the floor",
        they: "until it clips a knuckle and <ansi fg=\"yellow\">leaps from their hand</ansi>, skittering noisily across the floor",
    },
    {
        you:  "right up until your grip betrays you and it <ansi fg=\"yellow\">pings off the ground</ansi> with an embarrassing ring",
        they: "right up until their grip betrays them and it <ansi fg=\"yellow\">pings off the ground</ansi> with an embarrassing ring",
    },
    {
        you:  "then loses the battle with gravity and <ansi fg=\"yellow\">drops straight down</ansi> with a flat, final clack",
        they: "then loses the battle with gravity and <ansi fg=\"yellow\">drops straight down</ansi> with a flat, final clack",
    },
    {
        you:  "until a momentary lapse of concentration sends it <ansi fg=\"yellow\">tumbling to the floor</ansi> with a spin",
        they: "until a momentary lapse sends it <ansi fg=\"yellow\">tumbling from their fingers</ansi> with a spin",
    },
    {
        you:  "until you sneeze, completely unexpectedly, and the coin <ansi fg=\"yellow\">rockets off your hand</ansi> and bounces away",
        they: "until they sneeze without warning and the coin <ansi fg=\"yellow\">rockets off their hand</ansi> and bounces away",
    },
    {
        you:  "until a loud noise startles you and it <ansi fg=\"yellow\">leaps from your fingers</ansi> as if it had somewhere better to be",
        they: "until a loud noise makes them flinch and the coin <ansi fg=\"yellow\">leaps from their fingers</ansi> as if it had somewhere better to be",
    },
    {
        you:  "right up until you try to do a little extra flourish and the coin <ansi fg=\"yellow\">disagrees with your life choices</ansi> and hits the floor",
        they: "right up until they attempt a little extra flourish and the coin <ansi fg=\"yellow\">disagrees entirely</ansi> and hits the floor",
    },
];

function onCommand_roll(user, item, room) {

    var flavorIdx = UtilDiceRoll(1, ROLL_FLAVOR.length) - 1;
    var dropped = UtilDiceRoll(1, 10) === 1;

    if (dropped) {
        var dropIdx = UtilDiceRoll(1, DROP_FLAVOR.length) - 1;
        SendUserMessage(
            user.UserId(),
            "You roll your <ansi fg=\"item\">" + item.Name(true) + "</ansi> across your knuckles " + ROLL_FLAVOR[flavorIdx].you + ", " + DROP_FLAVOR[dropIdx].you + "."
        );
        SendRoomMessage(
            room.RoomId(),
            user.GetCharacterName(true) + " rolls their <ansi fg=\"item\">" + item.Name(true) + "</ansi> across their knuckles " + ROLL_FLAVOR[flavorIdx].they + ", " + DROP_FLAVOR[dropIdx].they + ".",
            user.UserId()
        );
        user.Command("drop " + item.ShorthandId());
    } else {
        SendUserMessage(
            user.UserId(),
            "You roll your <ansi fg=\"item\">" + item.Name(true) + "</ansi> across your knuckles " + ROLL_FLAVOR[flavorIdx].you + "."
        );
        SendRoomMessage(
            room.RoomId(),
            user.GetCharacterName(true) + " rolls their <ansi fg=\"item\">" + item.Name(true) + "</ansi> across their knuckles " + ROLL_FLAVOR[flavorIdx].they + ".",
            user.UserId()
        );
    }

    return true;
}
