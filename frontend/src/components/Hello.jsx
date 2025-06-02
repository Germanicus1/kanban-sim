import React from 'react';

export default function Hello({ name }) {
    return (
        <div className="p-4 bg-blue-100 rounded-lg shadow">
            <p className="text-lg">
                Hello, <strong>{name}</strong>! This is a React island.
            </p>
        </div>
    );
}
