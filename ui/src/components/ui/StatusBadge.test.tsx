/**
 * StatusBadge Component Tests
 *
 * Tests the StatusBadge component for correct rendering, accessibility,
 * and visual behavior across different status types and variants.
 */

import { cleanup, render, screen } from '@testing-library/react';
import { afterEach, describe, expect, it } from 'vitest';
import i18n from '../../i18n';
import { StatusBadge } from './StatusBadge';
import type { Status } from './StatusConfig';

describe('StatusBadge', () => {
  const statuses: Status[] = ['success', 'warning', 'error', 'unknown', 'loading'];

  describe('icon variant', () => {
    it('renders with success status', () => {
      render(<StatusBadge status="success" variant="icon" />);
      const badge = screen.getByRole('img');
      expect(badge).toBeInTheDocument();
      expect(badge).toHaveAccessibleName('Status: success');
    });

    it('renders with warning status', () => {
      render(<StatusBadge status="warning" variant="icon" />);
      const badge = screen.getByRole('img');
      expect(badge).toBeInTheDocument();
      expect(badge).toHaveAccessibleName('Status: warning');
    });

    it('renders with error status', () => {
      render(<StatusBadge status="error" variant="icon" />);
      const badge = screen.getByRole('img');
      expect(badge).toBeInTheDocument();
      expect(badge).toHaveAccessibleName('Status: error');
    });

    it('renders with unknown status', () => {
      render(<StatusBadge status="unknown" variant="icon" />);
      const badge = screen.getByRole('img');
      expect(badge).toBeInTheDocument();
      expect(badge).toHaveAccessibleName('Status: unknown');
    });

    it('renders with loading status', () => {
      render(<StatusBadge status="loading" variant="icon" />);
      const badge = screen.getByRole('img');
      expect(badge).toBeInTheDocument();
      expect(badge).toHaveAccessibleName('Status: loading');
    });

    it.each(statuses)('applies correct styling for %s status', (status) => {
      const { container } = render(<StatusBadge status={status} variant="icon" />);
      const badge = container.querySelector('[role="img"]');
      expect(badge).toBeInTheDocument();
    });
  });

  describe('dot variant', () => {
    it('renders a small dot indicator', () => {
      render(<StatusBadge status="success" variant="dot" />);
      const badge = screen.getByRole('img');
      expect(badge).toBeInTheDocument();
      expect(badge).toHaveAccessibleName('Status: success');
    });

    it.each(statuses)('renders dot for %s status', (status) => {
      render(<StatusBadge status={status} variant="dot" />);
      const badge = screen.getByRole('img');
      expect(badge).toBeInTheDocument();
    });
  });

  describe('sizes', () => {
    it('renders small size', () => {
      const { container } = render(<StatusBadge status="success" size="sm" />);
      const badge = container.querySelector('[role="img"]');
      expect(badge).toBeInTheDocument();
    });

    it('renders medium size', () => {
      const { container } = render(<StatusBadge status="success" size="md" />);
      const badge = container.querySelector('[role="img"]');
      expect(badge).toBeInTheDocument();
    });
  });

  describe('accessibility', () => {
    it('has role="img" for screen readers', () => {
      render(<StatusBadge status="error" />);
      expect(screen.getByRole('img')).toBeInTheDocument();
    });

    it('provides accessible name via aria-label', () => {
      render(<StatusBadge status="warning" />);
      const badge = screen.getByRole('img');
      expect(badge).toHaveAccessibleName('Status: warning');
    });

    it('hides decorative icon from screen readers', () => {
      const { container } = render(<StatusBadge status="success" variant="icon" />);
      const icon = container.querySelector('[aria-hidden="true"]');
      expect(icon).toBeInTheDocument();
    });
  });

  describe('locale', () => {
    afterEach(async () => {
      // Unmount first: switching back with the badge mounted re-renders it
      // outside act().
      cleanup();
      await i18n.changeLanguage('en');
    });

    // The accessible name is the only text a badge has, and it used to be an
    // English literal in StatusConfig that no i18n check could see (#1498).
    it.each([
      ['success', 'Estado: correcto'],
      ['warning', 'Estado: advertencia'],
      ['error', 'Estado: error'],
      ['unknown', 'Estado: desconocido'],
      ['loading', 'Estado: cargando'],
    ] as const)('names %s in Spanish', async (status, name) => {
      await i18n.changeLanguage('es');
      render(<StatusBadge status={status} variant="dot" />);

      expect(screen.getByRole('img')).toHaveAccessibleName(name);
    });
  });

  describe('custom className', () => {
    it('applies additional CSS classes', () => {
      const { container } = render(<StatusBadge status="success" className="custom-class" />);
      const badge = container.querySelector('.custom-class');
      expect(badge).toBeInTheDocument();
    });
  });
});
