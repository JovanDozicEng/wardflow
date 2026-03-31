/**
 * ExceptionsPage - Main page for exception management
 */

import { useState } from 'react';
import { ExceptionList } from '../components/ExceptionList';
import { ExceptionForm } from '../components/ExceptionForm';
import { FinalizeModal } from '../components/FinalizeModal';
import { CorrectionModal } from '../components/CorrectionModal';
import { Button } from '@/shared/components/ui/Button';
import { useExceptionActions } from '../hooks/useExceptionActions';
import type { ExceptionEvent } from '../types/exception.types';
import { Layout } from '@/shared/components/layout/Layout';
import { PageHeader } from '@/shared/components/layout/PageHeader';

export const ExceptionsPage = () => {
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [editingException, setEditingException] = useState<ExceptionEvent | null>(null);
  const [finalizingException, setFinalizingException] = useState<ExceptionEvent | null>(null);
  const [correctingException, setCorrectingException] = useState<ExceptionEvent | null>(null);
  
  const {
    createException,
    updateExceptionData,
    finalizeException,
    correctException,
    loading,
  } = useExceptionActions();

  const handleCreate = async (data: any) => {
    try {
      await createException(data);
      setShowCreateModal(false);
      console.log('Exception created successfully');
    } catch (err) {
      console.error('Failed to create exception:', err);
      throw err;
    }
  };

  const handleEdit = (exception: ExceptionEvent) => {
    setEditingException(exception);
  };

  const handleUpdate = async (data: { data: Record<string, any> }) => {
    if (!editingException) return;
    try {
      await updateExceptionData(editingException.id, data);
      setEditingException(null);
      console.log('Exception updated successfully');
    } catch (err) {
      console.error('Failed to update exception:', err);
      throw err;
    }
  };

  const handleFinalize = (exception: ExceptionEvent) => {
    setFinalizingException(exception);
  };

  const handleFinalizeSubmit = async () => {
    if (!finalizingException) return;
    try {
      await finalizeException(finalizingException.id);
      setFinalizingException(null);
      console.log('Exception finalized successfully');
    } catch (err) {
      console.error('Failed to finalize exception:', err);
      throw err;
    }
  };

  const handleCorrect = (exception: ExceptionEvent) => {
    setCorrectingException(exception);
  };

  const handleCorrectSubmit = async (data: { reason: string; data: Record<string, any> }) => {
    if (!correctingException) return;
    try {
      await correctException(correctingException.id, data);
      setCorrectingException(null);
      console.log('Correction created successfully');
    } catch (err) {
      console.error('Failed to create correction:', err);
      throw err;
    }
  };

  return (
    <Layout>
      <PageHeader
        title="Exceptions"
        subtitle="Track and manage exception events"
        actions={
          <Button
            variant="primary"
            onClick={() => setShowCreateModal(true)}
            className="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white"
          >
            + New Exception
          </Button>
        }
      />

      <div className="bg-blue-50 border border-blue-200 rounded-lg p-4 mb-6">
        <h3 className="font-medium text-blue-900 mb-2">Exception Workflow</h3>
        <ul className="text-sm text-blue-800 space-y-1">
          <li>• <strong>Draft:</strong> Exception can be edited and updated</li>
          <li>• <strong>Finalized:</strong> Exception is immutable, only corrections allowed</li>
          <li>• <strong>Corrected:</strong> A correction was made to a finalized exception</li>
        </ul>
      </div>

      <ExceptionList
        onEdit={handleEdit}
        onFinalize={handleFinalize}
        onCorrect={handleCorrect}
        actionLoading={loading}
      />

      <ExceptionForm
        isOpen={showCreateModal}
        onClose={() => setShowCreateModal(false)}
        onSubmit={handleCreate}
        loading={loading}
      />

      <ExceptionForm
        isOpen={editingException !== null}
        onClose={() => setEditingException(null)}
        onSubmit={handleUpdate}
        editMode
        exception={editingException || undefined}
        loading={loading}
      />

      <FinalizeModal
        isOpen={finalizingException !== null}
        onClose={() => setFinalizingException(null)}
        onSubmit={handleFinalizeSubmit}
        exception={finalizingException}
        loading={loading}
      />

      <CorrectionModal
        isOpen={correctingException !== null}
        onClose={() => setCorrectingException(null)}
        onSubmit={handleCorrectSubmit}
        exception={correctingException}
        loading={loading}
      />
    </Layout>
  );
};
