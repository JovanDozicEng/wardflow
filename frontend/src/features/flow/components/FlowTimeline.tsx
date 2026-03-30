/**
 * FlowTimeline Component
 * Displays patient flow state transitions in chronological order
 * Shows timestamp, actor, state changes, and reasons
 */

import { format } from 'date-fns';
import { ArrowRight, AlertTriangle, User, Bot } from 'lucide-react';
import type { FlowStateTransition } from '../types';
import { FlowStateLabels, FlowStateColors } from '../types';
import { Card } from '../../../shared/components/ui/Card';
import { Badge } from '../../../shared/components/ui/Badge';

interface FlowTimelineProps {
  transitions: FlowStateTransition[];
  currentState: string | null;
  isLoading?: boolean;
}

export const FlowTimeline = ({ transitions, currentState, isLoading }: FlowTimelineProps) => {
  if (isLoading) {
    return (
      <Card padding="md">
        <div className="animate-pulse space-y-4">
          {[1, 2, 3].map((i) => (
            <div key={i} className="flex gap-4">
              <div className="w-12 h-12 bg-gray-200 rounded-full" />
              <div className="flex-1 space-y-2">
                <div className="h-4 bg-gray-200 rounded w-1/4" />
                <div className="h-3 bg-gray-200 rounded w-1/2" />
              </div>
            </div>
          ))}
        </div>
      </Card>
    );
  }

  if (transitions.length === 0) {
    return (
      <Card padding="md">
        <div className="text-center py-8 text-gray-500">
          <p>No flow transitions recorded yet.</p>
          <p className="text-sm mt-2">The first transition will appear here once recorded.</p>
        </div>
      </Card>
    );
  }

  return (
    <div className="space-y-4">
      {/* Current State Header */}
      {currentState && (
        <Card padding="md" className="bg-blue-50 border-blue-200">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-gray-600 font-medium">Current State</p>
              <p className="text-lg font-semibold text-blue-900 mt-1">
                {FlowStateLabels[currentState as keyof typeof FlowStateLabels] || currentState}
              </p>
            </div>
            <StateIndicator state={currentState} size="lg" />
          </div>
        </Card>
      )}

      {/* Timeline */}
      <Card padding="none">
        <div className="divide-y divide-gray-200">
          {transitions.map((transition, index) => (
            <TimelineItem
              key={transition.id}
              transition={transition}
              isLatest={index === transitions.length - 1}
            />
          ))}
        </div>
      </Card>
    </div>
  );
};

interface TimelineItemProps {
  transition: FlowStateTransition;
  isLatest: boolean;
}

const TimelineItem = ({ transition, isLatest }: TimelineItemProps) => {
  return (
    <div className={`p-4 ${isLatest ? 'bg-blue-50' : ''}`}>
      <div className="flex gap-4">
        {/* Actor Icon */}
        <div className="flex-shrink-0">
          <div
            className={`w-10 h-10 rounded-full flex items-center justify-center ${
              transition.actorType === 'user' ? 'bg-blue-100' : 'bg-gray-100'
            }`}
          >
            {transition.actorType === 'user' ? (
              <User className="w-5 h-5 text-blue-600" />
            ) : (
              <Bot className="w-5 h-5 text-gray-600" />
            )}
          </div>
        </div>

        {/* Content */}
        <div className="flex-1 min-w-0">
          {/* Header: Timestamp and Actor */}
          <div className="flex items-start justify-between gap-2 mb-2">
            <div>
              <p className="text-sm font-medium text-gray-900">
                {format(new Date(transition.transitionedAt), 'MMM d, yyyy h:mm a')}
              </p>
              {transition.actorName && (
                <p className="text-xs text-gray-500 mt-0.5">{transition.actorName}</p>
              )}
            </div>
            {transition.isOverride && (
              <Badge variant="warning" className="flex items-center gap-1 bg-orange-100 text-orange-800">
                <AlertTriangle className="w-3 h-3" />
                Override
              </Badge>
            )}
          </div>

          {/* State Transition */}
          <div className="flex items-center gap-2 mb-2">
            <StateIndicator state={transition.fromState} />
            <ArrowRight className="w-4 h-4 text-gray-400" />
            <StateIndicator state={transition.toState} />
          </div>

          {/* Reason */}
          {transition.reason && (
            <div className="mt-2 p-2 bg-gray-50 rounded text-sm text-gray-700 border border-gray-200">
              <span className="font-medium">Reason:</span> {transition.reason}
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

interface StateIndicatorProps {
  state: string | null;
  size?: 'sm' | 'md' | 'lg';
}

const StateIndicator = ({ state, size = 'md' }: StateIndicatorProps) => {
  if (!state) {
    return (
      <span className="px-2 py-1 rounded text-xs bg-gray-100 text-gray-600">
        Initial
      </span>
    );
  }

  const label = FlowStateLabels[state as keyof typeof FlowStateLabels] || state;
  const colorKey = state as keyof typeof FlowStateColors;
  const color = FlowStateColors[colorKey] || 'gray';

  const colorClasses: Record<string, string> = {
    gray: 'bg-gray-100 text-gray-800 border-gray-300',
    yellow: 'bg-yellow-100 text-yellow-800 border-yellow-300',
    blue: 'bg-blue-100 text-blue-800 border-blue-300',
    purple: 'bg-purple-100 text-purple-800 border-purple-300',
    green: 'bg-green-100 text-green-800 border-green-300',
    orange: 'bg-orange-100 text-orange-800 border-orange-300',
    slate: 'bg-slate-100 text-slate-800 border-slate-300',
  };

  const sizeClasses = {
    sm: 'px-2 py-0.5 text-xs',
    md: 'px-2 py-1 text-sm',
    lg: 'px-3 py-2 text-base',
  };

  return (
    <span
      className={`inline-flex items-center rounded border font-medium ${colorClasses[color]} ${sizeClasses[size]}`}
    >
      {label}
    </span>
  );
};
