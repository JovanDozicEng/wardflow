/**
 * TransitionStateButton Component
 * Allows users to trigger flow state transitions
 * Shows valid next states and handles override workflow
 */

import { useState } from 'react';
import { ArrowRight, AlertTriangle, X } from 'lucide-react';
import type { FlowState, CreateTransitionRequest, OverrideTransitionRequest } from '../types';
import {
  FlowStateLabels,
  getNextValidStates,
  isValidTransition,
} from '../types';
import { Button } from '../../../shared/components/ui/Button';
import { Modal } from '../../../shared/components/ui/Modal';

interface TransitionStateButtonProps {
  currentState: FlowState | null;
  encounterId: string;
  onTransition: (data: CreateTransitionRequest) => Promise<void>;
  onOverride?: (data: OverrideTransitionRequest) => Promise<void>;
  canOverride?: boolean; // User has admin/operations role
  disabled?: boolean;
}

export const TransitionStateButton = ({
  currentState,
  encounterId: _encounterId,
  onTransition,
  onOverride,
  canOverride = false,
  disabled = false,
}: TransitionStateButtonProps) => {
  const [isOpen, setIsOpen] = useState(false);
  const [isOverrideMode, setIsOverrideMode] = useState(false);
  const [selectedState, setSelectedState] = useState<FlowState | null>(null);
  const [reason, setReason] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const validNextStates = getNextValidStates(currentState);
  const allStates: FlowState[] = [
    'arrived',
    'triage',
    'provider_eval',
    'diagnostics',
    'admitted',
    'discharge_ready',
    'discharged',
  ];

  const handleOpen = () => {
    setIsOpen(true);
    setIsOverrideMode(false);
    setSelectedState(null);
    setReason('');
    setError(null);
  };

  const handleClose = () => {
    if (!isSubmitting) {
      setIsOpen(false);
      setIsOverrideMode(false);
      setSelectedState(null);
      setReason('');
      setError(null);
    }
  };

  const handleSubmit = async () => {
    if (!selectedState) {
      setError('Please select a target state');
      return;
    }

    if (isOverrideMode && !reason.trim()) {
      setError('Reason is required for override transitions');
      return;
    }

    setIsSubmitting(true);
    setError(null);

    try {
      if (isOverrideMode && onOverride) {
        await onOverride({
          fromState: currentState,
          toState: selectedState,
          reason: reason.trim(),
        });
      } else {
        await onTransition({
          toState: selectedState,
          reason: reason.trim() || undefined,
        });
      }
      handleClose();
    } catch (err: any) {
      setError(err.response?.data?.error?.message || 'Failed to record transition');
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <>
      <Button onClick={handleOpen} disabled={disabled} className="flex items-center gap-2">
        <ArrowRight className="w-4 h-4" />
        Change State
      </Button>

      <Modal
        isOpen={isOpen}
        onClose={handleClose}
        title={isOverrideMode ? 'Override Flow State' : 'Change Flow State'}
      >
        <div className="space-y-4">
          {/* Current State Display */}
          {currentState && (
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Current State
              </label>
              <div className="px-3 py-2 bg-gray-100 rounded text-sm font-medium">
                {FlowStateLabels[currentState]}
              </div>
            </div>
          )}

          {/* State Selection */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              {isOverrideMode ? 'Target State (Override)' : 'Next State'}
            </label>
            <div className="space-y-2">
              {(isOverrideMode ? allStates : validNextStates).map((state) => {
                const isValid = isValidTransition(currentState, state);
                const isDisabled = !isOverrideMode && !isValid;

                return (
                  <button
                    key={state}
                    type="button"
                    disabled={isDisabled}
                    onClick={() => setSelectedState(state)}
                    className={`
                      w-full text-left px-4 py-3 rounded-lg border-2 transition-all
                      ${
                        selectedState === state
                          ? 'border-blue-500 bg-blue-50'
                          : 'border-gray-200 hover:border-gray-300'
                      }
                      ${isDisabled ? 'opacity-50 cursor-not-allowed' : 'cursor-pointer'}
                    `}
                  >
                    <div className="flex items-center justify-between">
                      <span className="font-medium">{FlowStateLabels[state]}</span>
                      {!isValid && isOverrideMode && (
                        <span className="text-xs text-orange-600 flex items-center gap-1">
                          <AlertTriangle className="w-3 h-3" />
                          Invalid transition
                        </span>
                      )}
                    </div>
                  </button>
                );
              })}
            </div>
          </div>

          {/* Reason Input */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Reason {isOverrideMode && <span className="text-red-500">*</span>}
            </label>
            <textarea
              value={reason}
              onChange={(e) => setReason(e.target.value)}
              rows={3}
              className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
              placeholder={
                isOverrideMode
                  ? 'Reason is required for override transitions'
                  : 'Optional reason for this transition'
              }
            />
          </div>

          {/* Error Display */}
          {error && (
            <div className="p-3 bg-red-50 border border-red-200 rounded text-sm text-red-700">
              {error}
            </div>
          )}

          {/* Actions */}
          <div className="flex items-center justify-between pt-2">
            <div>
              {canOverride && !isOverrideMode && (
                <button
                  type="button"
                  onClick={() => setIsOverrideMode(true)}
                  className="text-sm text-orange-600 hover:text-orange-700 font-medium flex items-center gap-1"
                >
                  <AlertTriangle className="w-4 h-4" />
                  Override Mode
                </button>
              )}
              {isOverrideMode && (
                <button
                  type="button"
                  onClick={() => {
                    setIsOverrideMode(false);
                    setSelectedState(null);
                  }}
                  className="text-sm text-gray-600 hover:text-gray-700 font-medium flex items-center gap-1"
                >
                  <X className="w-4 h-4" />
                  Cancel Override
                </button>
              )}
            </div>

            <div className="flex gap-2">
              <Button variant="secondary" onClick={handleClose} disabled={isSubmitting}>
                Cancel
              </Button>
              <Button
                onClick={handleSubmit}
                disabled={!selectedState || isSubmitting}
                isLoading={isSubmitting}
              >
                {isOverrideMode ? 'Override State' : 'Change State'}
              </Button>
            </div>
          </div>
        </div>
      </Modal>
    </>
  );
};
