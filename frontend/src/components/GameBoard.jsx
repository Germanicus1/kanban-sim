import React from "react";

export default function GameBoard({ columns, cards, moveCard }) {
    // REM: console.log("Game columns:", columns);
    // REM: console.log("Card columns:", cards.map(c => c.card_column));

    return (
        <div className="grid grid-cols-4 gap-4 mt-6">
            {columns.map((col) => (
                <div key={col} className="bg-white p-4 rounded shadow">
                    <h3 className="font-semibold text-lg mb-2">{col}</h3>
                    <div className="space-y-2">
                        {cards
                            .filter((card) => card.card_column === col)
                            .map((card) => {
                                const colIndex = columns.indexOf(card.card_column);
                                const prev = columns[colIndex - 1];
                                const next = columns[colIndex + 1];

                                return (
                                    <div key={card.id} className="bg-gray-100 p-2 rounded border text-sm">
                                        {card.title}
                                        <div className="mt-1 flex gap-2 text-xs">
                                            {prev && (
                                                <button
                                                    className="text-blue-600 underline"
                                                    onClick={() => moveCard(card.id, prev)}
                                                >
                                                    ← {prev}
                                                </button>
                                            )}
                                            {next && (
                                                <button
                                                    className="text-blue-600 underline"
                                                    onClick={() => moveCard(card.id, next)}
                                                >
                                                    {next} →
                                                </button>
                                            )}
                                        </div>
                                    </div>
                                );
                            })}
                    </div>
                </div>
            ))}
        </div>
    );
}
