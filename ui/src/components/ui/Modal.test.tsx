import { cleanup, fireEvent, render, screen } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { Modal, ModalBody, ModalFooter, ModalHeader } from './Modal';

afterEach(cleanup);

describe('Modal', () => {
  it('renders nothing while closed', () => {
    render(
      <Modal isOpen={false} onClose={vi.fn()} title="Delete profile">
        body
      </Modal>,
    );
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
  });

  it('names the dialog by its title, or by ariaLabel when it has none', () => {
    const { rerender } = render(
      <Modal isOpen onClose={vi.fn()} title="Delete profile">
        body
      </Modal>,
    );
    expect(screen.getByRole('dialog', { name: 'Delete profile' })).toBeInTheDocument();

    rerender(
      <Modal isOpen onClose={vi.fn()} ariaLabel="Confirm delete" showCloseButton={false}>
        body
      </Modal>,
    );
    expect(screen.getByRole('dialog', { name: 'Confirm delete' })).toBeInTheDocument();
    expect(screen.queryByRole('heading')).not.toBeInTheDocument();
  });

  it('closes on Escape pressed on a control inside the dialog (#1604)', () => {
    const onClose = vi.fn();
    render(
      <Modal isOpen onClose={onClose} title="Delete profile">
        <button type="button">inside</button>
      </Modal>,
    );
    fireEvent.keyDown(screen.getByRole('button', { name: 'inside' }), { key: 'Escape' });
    expect(onClose).toHaveBeenCalledTimes(1);
  });

  it('stays open on Escape when closeOnEscape is false', () => {
    const onClose = vi.fn();
    render(
      <Modal isOpen onClose={onClose} title="Delete profile" closeOnEscape={false}>
        <button type="button">inside</button>
      </Modal>,
    );
    fireEvent.keyDown(screen.getByRole('button', { name: 'inside' }), { key: 'Escape' });
    expect(onClose).not.toHaveBeenCalled();
  });

  it('closes from the backdrop only when closeOnBackdropClick allows it', () => {
    const onClose = vi.fn();
    const { rerender } = render(
      <Modal isOpen onClose={onClose} title="Delete profile" showCloseButton={false}>
        body
      </Modal>,
    );
    fireEvent.click(screen.getByRole('button'));
    expect(onClose).toHaveBeenCalledTimes(1);

    rerender(
      <Modal
        isOpen
        onClose={onClose}
        title="Delete profile"
        showCloseButton={false}
        closeOnBackdropClick={false}
      >
        body
      </Modal>,
    );
    expect(screen.queryByRole('button')).not.toBeInTheDocument();
  });

  it('locks page scroll while open and releases it on close', () => {
    const { rerender } = render(
      <Modal isOpen onClose={vi.fn()} title="Delete profile">
        body
      </Modal>,
    );
    expect(document.body.style.overflow).toBe('hidden');

    rerender(
      <Modal isOpen={false} onClose={vi.fn()} title="Delete profile">
        body
      </Modal>,
    );
    expect(document.body.style.overflow).toBe('');
  });

  it('lays out header, body and footer sections around their children', () => {
    render(
      <Modal isOpen onClose={vi.fn()} title="Delete profile">
        <ModalHeader className="extra-header">intro</ModalHeader>
        <ModalBody>fields</ModalBody>
        <ModalFooter>actions</ModalFooter>
      </Modal>,
    );
    expect(screen.getByText('intro')).toHaveClass('mb-content', 'extra-header');
    expect(screen.getByText('fields')).toHaveClass('stack-lg');
    expect(screen.getByText('actions')).toHaveClass('justify-end', 'border-t');
  });
});
