/**
 * CollapsibleSection Component Tests
 *
 * Tests the CollapsibleSection component for correct rendering, toggling,
 * accessibility, and status/count badge display.
 */

import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { CollapsibleSection } from './CollapsibleSection';

describe('CollapsibleSection', () => {
  describe('rendering', () => {
    it('renders title', () => {
      render(<CollapsibleSection title="Test Section">Content</CollapsibleSection>);
      expect(screen.getByText('Test Section')).toBeInTheDocument();
    });

    it('does not render children when closed', () => {
      render(
        <CollapsibleSection title="Test Section" defaultOpen={false}>
          <p>Hidden Content</p>
        </CollapsibleSection>,
      );
      expect(screen.queryByText('Hidden Content')).not.toBeInTheDocument();
    });

    it('renders children when defaultOpen is true', () => {
      render(
        <CollapsibleSection title="Test Section" defaultOpen={true}>
          <p>Visible Content</p>
        </CollapsibleSection>,
      );
      expect(screen.getByText('Visible Content')).toBeInTheDocument();
    });
  });

  describe('count badge', () => {
    it('displays count in header', () => {
      render(
        <CollapsibleSection title="Items" count={5}>
          Content
        </CollapsibleSection>,
      );
      expect(screen.getByText('(5)')).toBeInTheDocument();
    });

    it('displays zero count', () => {
      render(
        <CollapsibleSection title="Items" count={0}>
          Content
        </CollapsibleSection>,
      );
      expect(screen.getByText('(0)')).toBeInTheDocument();
    });

    it('does not display count when not provided', () => {
      render(<CollapsibleSection title="No Count">Content</CollapsibleSection>);
      expect(screen.queryByText('(')).not.toBeInTheDocument();
    });
  });

  describe('status indicator', () => {
    it('displays status badge when status is provided', () => {
      render(
        <CollapsibleSection title="Status Section" status="success">
          Content
        </CollapsibleSection>,
      );
      const badge = screen.getByRole('img', { name: /status: success/i });
      expect(badge).toBeInTheDocument();
    });

    it('displays warning status', () => {
      render(
        <CollapsibleSection title="Warning Section" status="warning">
          Content
        </CollapsibleSection>,
      );
      const badge = screen.getByRole('img', { name: /status: warning/i });
      expect(badge).toBeInTheDocument();
    });

    it('displays error status', () => {
      render(
        <CollapsibleSection title="Error Section" status="error">
          Content
        </CollapsibleSection>,
      );
      const badge = screen.getByRole('img', { name: /status: error/i });
      expect(badge).toBeInTheDocument();
    });
  });

  describe('toggling', () => {
    it('expands when header is clicked', () => {
      render(
        <CollapsibleSection title="Click Me" defaultOpen={false}>
          <p>Expandable Content</p>
        </CollapsibleSection>,
      );

      expect(screen.queryByText('Expandable Content')).not.toBeInTheDocument();

      const button = screen.getByRole('button', { name: /click me/i });
      fireEvent.click(button);

      expect(screen.getByText('Expandable Content')).toBeInTheDocument();
    });

    it('collapses when header is clicked again', () => {
      render(
        <CollapsibleSection title="Toggle Me" defaultOpen={true}>
          <p>Collapsible Content</p>
        </CollapsibleSection>,
      );

      expect(screen.getByText('Collapsible Content')).toBeInTheDocument();

      const button = screen.getByRole('button', { name: /toggle me/i });
      fireEvent.click(button);

      expect(screen.queryByText('Collapsible Content')).not.toBeInTheDocument();
    });
  });

  describe('variants', () => {
    it('applies default variant styling', () => {
      const { container } = render(
        <CollapsibleSection title="Default" variant="default">
          Content
        </CollapsibleSection>,
      );
      const section = container.querySelector('section');
      expect(section?.className).toContain('overflow-hidden');
    });

    it('applies compact variant styling', () => {
      const { container } = render(
        <CollapsibleSection title="Compact" variant="compact">
          Content
        </CollapsibleSection>,
      );
      const section = container.querySelector('section');
      expect(section?.className).not.toContain('overflow-hidden');
    });
  });

  describe('accessibility', () => {
    it('has aria-expanded attribute', () => {
      render(
        <CollapsibleSection title="Accessible" defaultOpen={false}>
          Content
        </CollapsibleSection>,
      );

      const button = screen.getByRole('button');
      expect(button).toHaveAttribute('aria-expanded', 'false');
    });

    it('updates aria-expanded when toggled', () => {
      render(
        <CollapsibleSection title="Accessible" defaultOpen={false}>
          Content
        </CollapsibleSection>,
      );

      const button = screen.getByRole('button');
      expect(button).toHaveAttribute('aria-expanded', 'false');

      fireEvent.click(button);
      expect(button).toHaveAttribute('aria-expanded', 'true');
    });

    it('uses semantic section element', () => {
      const { container } = render(
        <CollapsibleSection title="Semantic">Content</CollapsibleSection>,
      );
      expect(container.querySelector('section')).toBeInTheDocument();
    });
  });

  describe('custom className', () => {
    it('applies additional CSS classes', () => {
      const { container } = render(
        <CollapsibleSection title="Custom" className="custom-class">
          Content
        </CollapsibleSection>,
      );
      expect(container.querySelector('.custom-class')).toBeInTheDocument();
    });
  });
});
