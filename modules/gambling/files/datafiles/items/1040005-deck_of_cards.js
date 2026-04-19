
var SUITS = [
    { name: "Spades",   symbol: "♠", color: "black-bold" },
    { name: "Hearts",   symbol: "♥", color: "red-bold"   },
    { name: "Diamonds", symbol: "♦", color: "red-bold"   },
    { name: "Clubs",    symbol: "♣", color: "black-bold" },
];

var VALUES = [
    { name: "Ace",   short: "A"  },
    { name: "Two",   short: "2"  },
    { name: "Three", short: "3"  },
    { name: "Four",  short: "4"  },
    { name: "Five",  short: "5"  },
    { name: "Six",   short: "6"  },
    { name: "Seven", short: "7"  },
    { name: "Eight", short: "8"  },
    { name: "Nine",  short: "9"  },
    { name: "Ten",   short: "10" },
    { name: "Jack",  short: "J"  },
    { name: "Queen", short: "Q"  },
    { name: "King",  short: "K"  },
];

var SHUFFLE_MESSAGES = [
    "You riffle the <ansi fg=\"item\">{item}</ansi> with a crisp snap, the cards blurring together.",
    "You give the <ansi fg=\"item\">{item}</ansi> a slow overhand shuffle, squaring the deck after each pass.",
    "You bridge the <ansi fg=\"item\">{item}</ansi> and let the cards cascade together with a satisfying flutter.",
    "You cut the <ansi fg=\"item\">{item}</ansi> three times and shuffle twice, an old habit.",
    "You shuffle the <ansi fg=\"item\">{item}</ansi> absently, the motion as natural as breathing.",
];

var DRAW_INTROS = [
    "You cut the deck and draw from the top.",
    "You draw from the middle of the deck.",
    "You fan the cards and pull one without looking.",
    "You draw the bottom card.",
    "You let your hand hover over the spread deck until one card feels right.",
];

function randomCard() {
    var suit  = SUITS[UtilDiceRoll(1, SUITS.length) - 1];
    var value = VALUES[UtilDiceRoll(1, VALUES.length) - 1];
    return {
        display: "<ansi fg=\"" + suit.color + "\">" + value.short + suit.symbol + "</ansi>",
        full:    value.name + " of " + suit.name,
    };
}

function onCommand_shuffle(user, item, room) {

    var msg = SHUFFLE_MESSAGES[UtilDiceRoll(1, SHUFFLE_MESSAGES.length) - 1]
        .replace("{item}", item.Name(true));

    SendUserMessage(user.UserId(), msg);
    SendRoomMessage(
        room.RoomId(),
        user.GetCharacterName(true) + " shuffles their <ansi fg=\"item\">" + item.Name(true) + "</ansi>.",
        user.UserId()
    );

    return true;
}

function onCommand_draw(user, item, room) {

    var card  = randomCard();
    var intro = DRAW_INTROS[UtilDiceRoll(1, DRAW_INTROS.length) - 1];

    SendUserMessage(
        user.UserId(),
        intro + " You draw the <ansi fg=\"white-bold\">" + card.full + "</ansi>  " + card.display
    );

    SendRoomMessage(
        room.RoomId(),
        user.GetCharacterName(true) + " draws a card from their <ansi fg=\"item\">" + item.Name(true) + "</ansi>: " + card.display,
        user.UserId()
    );

    return true;
}

function onCommand_deal(user, item, room) {

    var players = room.GetPlayers();

    if (players.length === 0) {
        SendUserMessage(user.UserId(), "There is no one here to deal to.");
        return true;
    }

    // Build and shuffle a full deck, then deal from the top.
    var deck = [];
    for (var s = 0; s < SUITS.length; s++) {
        for (var v = 0; v < VALUES.length; v++) {
            deck.push({ suit: SUITS[s], value: VALUES[v] });
        }
    }
    // Fisher-Yates shuffle
    for (var i = deck.length - 1; i > 0; i--) {
        var j = UtilDiceRoll(1, i + 1) - 1;
        var tmp = deck[i];
        deck[i] = deck[j];
        deck[j] = tmp;
    }

    SendRoomMessage(
        room.RoomId(),
        user.GetCharacterName(true) + " deals a card to everyone from their <ansi fg=\"item\">" + item.Name(true) + "</ansi>."
    );

    for (var i = 0; i < players.length; i++) {
        var entry = deck[i];
        var display = "<ansi fg=\"" + entry.suit.color + "\">" + entry.value.short + entry.suit.symbol + "</ansi>";
        var full    = entry.value.name + " of " + entry.suit.name;

        SendUserMessage(
            players[i].UserId(),
            user.GetCharacterName(true) + " deals you the <ansi fg=\"white-bold\">" + full + "</ansi>  " + display
        );

        SendRoomMessage(
            room.RoomId(),
            user.GetCharacterName(true) + " deals " + players[i].GetCharacterName(true) + " the " + display,
            players[i].UserId()
        );
    }

    return true;
}
