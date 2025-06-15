import React, { useState, useEffect } from 'react';
import {
  generateHouseholdJoinCode,
  joinHousehold,
  leaveHousehold,
  removeHouseholdMember,
  getHousehold,
  HouseholdJoinCode,
  Household,
} from '../api';

const HouseholdManager: React.FC = () => {
  const [joinCode, setJoinCode] = useState<HouseholdJoinCode | null>(null);
  const [household, setHousehold] = useState<Household | null>(null);
  const [joinInput, setJoinInput] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  const fetchHousehold = async () => {
    try {
      const res = await getHousehold();
      setHousehold(res.data);
    } catch (err: any) {
      setError('Failed to fetch household info.');
    }
  };

  useEffect(() => {
    fetchHousehold();
  }, []);

  const handleGenerateJoinCode = async () => {
    setLoading(true);
    setError(null);
    setSuccess(null);
    try {
      const res = await generateHouseholdJoinCode();
      setJoinCode(res.data);
      setSuccess('Join code generated!');
    } catch (err: any) {
      setError('Failed to generate join code.');
    }
    setLoading(false);
  };

  const handleJoinHousehold = async () => {
    setLoading(true);
    setError(null);
    setSuccess(null);
    try {
      await joinHousehold(joinInput);
      setSuccess('Joined household!');
      setJoinInput('');
      fetchHousehold();
    } catch (err: any) {
      setError('Failed to join household.');
    }
    setLoading(false);
  };

  const handleLeaveHousehold = async () => {
    setLoading(true);
    setError(null);
    setSuccess(null);
    try {
      await leaveHousehold();
      setSuccess('Left household.');
      fetchHousehold();
    } catch (err: any) {
      setError('Failed to leave household.');
    }
    setLoading(false);
  };

  const handleRemoveMember = async (userId: string) => {
    setLoading(true);
    setError(null);
    setSuccess(null);
    try {
      await removeHouseholdMember(userId);
      setSuccess('Member removed.');
      fetchHousehold();
    } catch (err: any) {
      setError('Failed to remove member.');
    }
    setLoading(false);
  };

  return (
    <div className="bg-base-100 p-6 rounded-lg shadow max-w-lg mx-auto mt-8">
      <h2 className="text-2xl font-bold mb-4">Household Management</h2>
      {household && (
        <div className="mb-4">
          <div className="font-semibold">
            Household Name:{' '}
            <span className="badge badge-neutral">{household.name}</span>
          </div>
          <div className="text-xs text-base-content/60">
            Household ID: {household.id}
          </div>
        </div>
      )}
      {error && <div className="alert alert-error mb-4">{error}</div>}
      {success && <div className="alert alert-success mb-4">{success}</div>}

      <div className="mb-6">
        <button
          className="btn btn-primary"
          onClick={handleGenerateJoinCode}
          disabled={loading}
        >
          Generate Join Code
        </button>
        {joinCode && (
          <div className="mt-2">
            <span className="badge badge-info">Code: {joinCode.code}</span>
            <span className="ml-2 text-xs text-base-content/60">
              Household ID: {joinCode.household_id}
            </span>
            <span className="ml-2 text-xs text-base-content/60">
              Expires: {new Date(joinCode.expires_at).toLocaleString()}
            </span>
          </div>
        )}
      </div>

      <div className="mb-6">
        <input
          className="input input-bordered mr-2"
          placeholder="Enter join code"
          value={joinInput}
          onChange={(e) => setJoinInput(e.target.value)}
          disabled={loading}
        />
        <button
          className="btn btn-secondary"
          onClick={handleJoinHousehold}
          disabled={loading || !joinInput}
        >
          Join Household
        </button>
      </div>

      <div className="mb-6">
        <button
          className="btn btn-warning"
          onClick={handleLeaveHousehold}
          disabled={loading}
        >
          Leave Household
        </button>
      </div>

      <div>
        <h3 className="font-semibold mb-2">Members:</h3>
        <ul className="list-disc pl-6">
          {household?.members.map((member) => (
            <li key={member.user_id} className="flex items-center justify-between mb-1">
              <span>{member.email}</span>
              <button
                className="btn btn-xs btn-error ml-2"
                onClick={() => handleRemoveMember(member.user_id)}
                disabled={loading}
              >
                Remove
              </button>
            </li>
          ))}
        </ul>
      </div>
    </div>
  );
};

export default HouseholdManager;
