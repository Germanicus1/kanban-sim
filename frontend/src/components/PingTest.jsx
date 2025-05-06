import { useEffect, useState } from "react";
import GameBoard from "./GameBoard";
import { supabase } from "../supabase";

const useQuery = () => new URLSearchParams(window.location.search);


export default function PingTest() {
    const [message, setMessage] = useState("Loading...");
    const [game, setGame] = useState(null);
    const [cards, setCards] = useState([]);
    const [players, setPlayers] = useState([]);


    useEffect(() => {
        fetch("http://localhost:8080/ping")
            .then((res) => res.text())
            .then(setMessage)
            .catch(() => setMessage("Error contacting backend"));
    }, []);

    useEffect(() => {
        if (typeof window === "undefined") return;
        const params = new URLSearchParams(window.location.search);
        const queryGameId = params.get("game");
        if (queryGameId) {
            fetch(`http://localhost:8080/game/${queryGameId}`)
                .then(res => res.json())
                .then((data) => {
                    const loadedGame = data[0];
                    setGame(loadedGame);
                    return Promise.all([
                        fetch(`http://localhost:8080/cards/${loadedGame.id}`),
                        fetch(`http://localhost:8080/game/${loadedGame.id}/players`)
                    ]);
                })
                .then(async ([cardsRes, playersRes]) => {
                    const cardData = await cardsRes.json();
                    const playerData = await playersRes.json();
                    setCards(cardData);
                    setPlayers(playerData);
                })
                .catch(err => console.error("Failed to auto-load game from URL:", err));
        }
    }, []);


    useEffect(() => {
        if (!game?.id) return;

        const playerChannel = supabase
            .channel(`players:${game.id}`)
            .on('postgres_changes', {
                event: '*',
                schema: 'public',
                table: 'players',
                filter: `game_id=eq.${game.id}`
            }, () => {
                fetch(`http://localhost:8080/game/${game.id}/players`)
                    .then(res => res.json())
                    .then(setPlayers);
            })
            .subscribe();

        const cardChannel = supabase
            .channel(`cards:${game.id}`)
            .on('postgres_changes', {
                event: '*',
                schema: 'public',
                table: 'cards',
                filter: `game_id=eq.${game.id}`
            }, () => {
                fetch(`http://localhost:8080/cards/${game.id}`)
                    .then(res => res.json())
                    .then(setCards);
            })
            .subscribe();

        const eventChannel = supabase
            .channel(`game_events:${game.id}`)
            .on('postgres_changes', {
                event: 'INSERT',
                schema: 'public',
                table: 'game_events',
                filter: `game_id=eq.${game.id}`
            }, (payload) => {
                if (payload.new?.type === "ended") {
                    alert("This game has ended.");
                    setGame(null);
                    setCards([]);
                    setPlayers([]);
                    localStorage.removeItem("playerId");
                    localStorage.removeItem("playerName");
                }
            })
            .subscribe();

        return () => {
            supabase.removeChannel(playerChannel);
            supabase.removeChannel(cardChannel);
            supabase.removeChannel(eventChannel);
        };
    }, [game?.id]);

    const createGame = async () => {
        const res = await fetch("http://localhost:8080/create-game");
        const [game] = await res.json();
        setGame(game);
        const [cards, players] = await Promise.all([
            fetch(`http://localhost:8080/cards/${game.id}`).then(res => res.json()),
            fetch(`http://localhost:8080/game/${game.id}/players`).then(res => res.json())
        ]);
        setCards(cards);
        setPlayers(players);
    };

    const loadGame = async () => {
        const gameId = prompt("Enter game ID:");
        const [game] = await fetch(`http://localhost:8080/game/${gameId}`).then(res => res.json());
        setGame(game);
        const [cards, players] = await Promise.all([
            fetch(`http://localhost:8080/cards/${game.id}`).then(res => res.json()),
            fetch(`http://localhost:8080/game/${game.id}/players`).then(res => res.json())
        ]);
        setCards(cards);
        setPlayers(players);
        const storedId = localStorage.getItem("playerId");
        const found = players.find(p => p.id === storedId);
        if (found) console.log("Auto-rejoined as:", found.name);
    };

    const moveCard = async (cardId, nextColumn) => {
        await fetch(`http://localhost:8080/cards/${cardId}/move`, {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ new_column: nextColumn }),
        });
    };

    const nextDay = async () => {
        const res = await fetch(`http://localhost:8080/game/${game.id}/next-day`, { method: "POST" });
        const updated = await res.json();
        setGame(updated[0]);
    };

    const joinGame = async () => {
        const name = prompt("Enter your name:");
        if (!name || !game?.id) return;
        const res = await fetch(`http://localhost:8080/game/${game.id}/join`, {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ name }),
        });
        if (res.status === 409) return alert("Name already taken.");
        const [player] = await res.json();
        localStorage.setItem("playerId", player.id);
        localStorage.setItem("playerName", player.name);
        alert(`Welcome, ${player.name}! Player ID: ${player.id}`);
    };

    const leaveGame = async () => {
        const playerId = localStorage.getItem("playerId");
        if (!playerId || !game?.id) return;
        const res = await fetch(`http://localhost:8080/game/${game.id}/leave`, {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ player_id: playerId }),
        });
        if (res.ok) {
            localStorage.removeItem("playerId");
            localStorage.removeItem("playerName");
            alert("You have left the game.");
            setGame(null);
        }
    };

    const resetGame = async () => {
        if (!game?.id) return;
        const confirmed = confirm("Reset game?");
        if (!confirmed) return;
        await fetch(`http://localhost:8080/game/${game.id}/reset`, { method: "POST" });
    };

    const endGame = async () => {
        if (!game?.id) return;
        const confirmed = confirm("End game and delete all data?");
        if (!confirmed) return;
        await fetch(`http://localhost:8080/game/${game.id}/end`, { method: "POST" });
        alert("Game ended and deleted.");
        setGame(null);
        setCards([]);
        setPlayers([]);
        localStorage.removeItem("playerId");
        localStorage.removeItem("playerName");
    };

    return (
        <div className="mt-6 p-4 border rounded bg-white shadow">
            <h2 className="text-xl font-semibold mb-2">Backend Status</h2>
            <p>{message}</p>

            <div className="space-x-2 mt-4">
                <button onClick={createGame} className="bg-blue-600 text-white px-4 py-2 rounded">Create Game</button>
                <button onClick={loadGame} className="bg-gray-600 text-white px-4 py-2 rounded">Load Game</button>
                <button onClick={nextDay} className="bg-green-600 text-white px-4 py-2 rounded">Next Day</button>
                <button onClick={joinGame} className="bg-purple-600 text-white px-4 py-2 rounded">Join Game</button>
                <button onClick={resetGame} className="bg-yellow-600 text-white px-4 py-2 rounded">Reset Game</button>
                <button onClick={leaveGame} className="bg-red-600 text-white px-4 py-2 rounded">Leave Game</button>
                <button onClick={endGame} className="bg-black text-white px-4 py-2 rounded">End Game</button>
            </div>

            {game && (
                <div className="mt-4">
                    <p><strong>Game ID:</strong> {game.id}</p>
                    <p><strong>Day:</strong> {game.day}</p>
                    {players.length > 0 && (
                        <ul className="list-disc list-inside">
                            {players.map((p) => (
                                <li key={p.id}>{p.name} {p.id === localStorage.getItem("playerId") ? "(You)" : ""}</li>
                            ))}
                        </ul>
                    )}
                    <p><strong>Columns:</strong> {game.columns?.join(" → ") ?? "(no columns)"}</p>
                    <GameBoard columns={game.columns} cards={cards} moveCard={moveCard} />
                </div>
            )}
        </div>
    );
}
