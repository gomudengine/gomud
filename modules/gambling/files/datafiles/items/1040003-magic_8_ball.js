
// Classic 8-ball responses, grouped by sentiment for flavor
var RESPONSES_YES = [
    { text: "It is certain.",          color: "green" },
    { text: "It is decidedly so.",     color: "green" },
    { text: "Without a doubt.",        color: "green" },
    { text: "Yes, definitely.",        color: "green" },
    { text: "You may rely on it.",     color: "green" },
    { text: "As I see it, yes.",       color: "green" },
    { text: "Most likely.",            color: "green" },
    { text: "Outlook good.",           color: "green" },
    { text: "Yes.",                    color: "green" },
    { text: "Signs point to yes.",     color: "green" },
];

var RESPONSES_MAYBE = [
    { text: "Reply hazy, try again.",        color: "yellow" },
    { text: "Ask again later.",              color: "yellow" },
    { text: "Better not tell you now.",      color: "yellow" },
    { text: "Cannot predict now.",           color: "yellow" },
    { text: "Concentrate and ask again.",    color: "yellow" },
];

var RESPONSES_NO = [
    { text: "Don't count on it.",      color: "red" },
    { text: "My reply is no.",         color: "red" },
    { text: "My sources say no.",      color: "red" },
    { text: "Outlook not so good.",    color: "red" },
    { text: "Very doubtful.",          color: "red" },
];

var SHAKE_MESSAGES = [
    "You turn the <ansi fg=\"item\">{item}</ansi> over in your hands, watching the dark liquid swirl.",
    "You give the <ansi fg=\"item\">{item}</ansi> a slow, deliberate shake.",
    "You cup the <ansi fg=\"item\">{item}</ansi> in both hands and give it a firm shake.",
    "You shake the <ansi fg=\"item\">{item}</ansi> vigorously, perhaps more than is strictly necessary.",
    "You hold the <ansi fg=\"item\">{item}</ansi> up, give it a single shake, and peer into the window.",
];

function onCommand_shake(user, item, room) {

    // Pick a response pool weighted: 10 yes, 5 maybe, 5 no
    var all = RESPONSES_YES.concat(RESPONSES_MAYBE).concat(RESPONSES_NO);
    var response = all[UtilDiceRoll(1, all.length) - 1];

    var shakeIdx = UtilDiceRoll(1, SHAKE_MESSAGES.length) - 1;
    var shakeMsg = SHAKE_MESSAGES[shakeIdx].replace("{item}", item.Name(true));

    SendUserMessage(
        user.UserId(),
        shakeMsg + "\n" +
        "  The answer floats up from the depths: <ansi fg=\"" + response.color + "-bold\">\"" + response.text + "\"</ansi>"
    );

    SendRoomMessage(
        room.RoomId(),
        user.GetCharacterName(true) + " shakes their <ansi fg=\"item\">" + item.Name(true) + "</ansi> and peers into it intently.",
        user.UserId()
    );

    return true;
}
