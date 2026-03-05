import { useState } from 'react';

interface AuthScreenProps {
  onAuthenticated: () => void;
}

export function AuthScreen({ onAuthenticated }: AuthScreenProps) {
  const [token, setToken] = useState('');

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (token.trim()) {
      localStorage.setItem('api_token', token.trim());
      onAuthenticated();
    }
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-gray-50">
      <form
        onSubmit={handleSubmit}
        className="w-full max-w-sm rounded-lg bg-white p-6 shadow-md"
      >
        <h1 className="text-lg font-bold text-gray-900 mb-4">
          API Authentication
        </h1>
        <p className="text-sm text-gray-500 mb-4">
          Enter your API bearer token to continue.
        </p>
        <input
          type="password"
          value={token}
          onChange={(e) => setToken(e.target.value)}
          placeholder="Bearer token"
          className="w-full rounded border border-gray-300 px-3 py-2 text-sm mb-3 focus:outline-none focus:ring-2 focus:ring-blue-500"
        />
        <button
          type="submit"
          className="w-full rounded bg-blue-500 px-3 py-2 text-sm font-medium text-white hover:bg-blue-600"
        >
          Connect
        </button>
      </form>
    </div>
  );
}
