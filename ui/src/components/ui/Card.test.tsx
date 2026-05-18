/**
 * Card Component Tests
 *
 * Tests the Card component and its sub-components (CardValue, CardRow, CardDivider)
 * for correct rendering, accessibility, and user interaction.
 */

import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { Card, CardDivider, CardRow, CardValue } from './Card';

describe('Card', () => {
  describe('rendering', () => {
    it('renders title', () => {
      render(
        <Card title="Test Card" status="success">
          Content
        </Card>,
      );
      expect(screen.getByText('Test Card')).toBeInTheDocument();
    });

    it('renders subtitle when provided', () => {
      render(
        <Card title="Test Card" subtitle="Test subtitle" status="success">
          Content
        </Card>,
      );
      expect(screen.getByText('Test subtitle')).toBeInTheDocument();
    });

    it('renders children content', () => {
      render(
        <Card title="Test Card" status="success">
          <p>Child content</p>
        </Card>,
      );
      expect(screen.getByText('Child content')).toBeInTheDocument();
    });

    it('renders icon when provided', () => {
      render(
        <Card title="Test Card" status="success" icon={<span data-testid="icon">Icon</span>}>
          Content
        </Card>,
      );
      expect(screen.getByTestId('icon')).toBeInTheDocument();
    });

    it('renders headerAction when provided', () => {
      render(
        <Card
          title="Test Card"
          status="success"
          headerAction={<button type="button">Action</button>}
        >
          Content
        </Card>,
      );
      expect(screen.getByRole('button', { name: 'Action' })).toBeInTheDocument();
    });

    it('applies custom className', () => {
      const { container } = render(
        <Card title="Test Card" status="success" className="custom-class">
          Content
        </Card>,
      );
      expect(container.querySelector('.custom-class')).toBeInTheDocument();
    });
  });

  describe('interaction', () => {
    it('calls onClick when clicked', () => {
      const handleClick = vi.fn();
      render(
        <Card title="Clickable Card" status="success" onClick={handleClick}>
          Content
        </Card>,
      );

      const card = screen.getByRole('button');
      fireEvent.click(card);
      expect(handleClick).toHaveBeenCalledTimes(1);
    });

    it('has button role when onClick is provided', () => {
      render(
        <Card
          title="Interactive Card"
          status="success"
          onClick={(): void => {
            /* noop for test */
          }}
        >
          Content
        </Card>,
      );
      expect(screen.getByRole('button')).toBeInTheDocument();
    });

    it('does not have button role when onClick is not provided', () => {
      render(
        <Card title="Static Card" status="success">
          Content
        </Card>,
      );
      expect(screen.queryByRole('button')).not.toBeInTheDocument();
    });

    it('handles Enter key activation', () => {
      const handleClick = vi.fn();
      render(
        <Card title="Keyboard Card" status="success" onClick={handleClick}>
          Content
        </Card>,
      );

      const card = screen.getByRole('button');
      fireEvent.keyDown(card, { key: 'Enter' });
      expect(handleClick).toHaveBeenCalledTimes(1);
    });

    it('handles Space key activation', () => {
      const handleClick = vi.fn();
      render(
        <Card title="Keyboard Card" status="success" onClick={handleClick}>
          Content
        </Card>,
      );

      const card = screen.getByRole('button');
      fireEvent.keyDown(card, { key: ' ' });
      expect(handleClick).toHaveBeenCalledTimes(1);
    });
  });

  describe('accessibility', () => {
    it('has aria-label by default', () => {
      const { container } = render(
        <Card title="Accessible Card" status="success">
          Content
        </Card>,
      );
      const card = container.firstChild;
      expect(card).toHaveAttribute('aria-label', 'Accessible Card');
    });

    it('combines title and subtitle for aria-label', () => {
      const { container } = render(
        <Card title="Card Title" subtitle="Card Subtitle" status="success">
          Content
        </Card>,
      );
      const card = container.firstChild;
      expect(card).toHaveAttribute('aria-label', 'Card Title - Card Subtitle');
    });

    it('uses custom ariaLabel when provided', () => {
      const { container } = render(
        <Card title="Card" status="success" ariaLabel="Custom label">
          Content
        </Card>,
      );
      const card = container.firstChild;
      expect(card).toHaveAttribute('aria-label', 'Custom label');
    });

    it('has tabIndex 0 for interactive cards', () => {
      const { container } = render(
        <Card
          title="Card"
          status="success"
          onClick={(): void => {
            /* noop for test */
          }}
        >
          Content
        </Card>,
      );
      const card = container.firstChild;
      expect(card).toHaveAttribute('tabindex', '0');
    });
  });

  describe('live region', () => {
    it('adds aria-live when enableLiveRegion is true', () => {
      const { container } = render(
        <Card title="Live Card" status="success" enableLiveRegion={true}>
          Content
        </Card>,
      );
      const card = container.firstChild;
      expect(card).toHaveAttribute('aria-live', 'polite');
      expect(card).toHaveAttribute('aria-atomic', 'true');
    });

    it('does not add aria-live when enableLiveRegion is false', () => {
      const { container } = render(
        <Card title="Static Card" status="success">
          Content
        </Card>,
      );
      const card = container.firstChild;
      expect(card).not.toHaveAttribute('aria-live');
    });
  });
});

describe('CardValue', () => {
  it('renders value', () => {
    render(<CardValue value="42" />);
    expect(screen.getByTestId('card-value')).toHaveTextContent('42');
  });

  it('renders label when provided', () => {
    render(<CardValue label="Count" value="42" />);
    expect(screen.getByText('Count')).toBeInTheDocument();
  });

  it('renders unit when provided', () => {
    render(<CardValue value="100" unit="ms" />);
    expect(screen.getByText('ms')).toBeInTheDocument();
  });

  it('applies monospace font when mono is true', () => {
    const { container } = render(<CardValue value="12345" mono={true} />);
    const valueEl = container.querySelector('.font-mono');
    expect(valueEl).toBeInTheDocument();
  });

  it('allows text wrapping when allowWrap is true', () => {
    const { container } = render(<CardValue value="Long value" allowWrap={true} />);
    const valueEl = container.querySelector('.break-all');
    expect(valueEl).toBeInTheDocument();
  });
});

describe('CardRow', () => {
  it('renders label and value', () => {
    render(<CardRow label="Label" value="Value" />);
    expect(screen.getByText('Label')).toBeInTheDocument();
    expect(screen.getByTestId('card-row-value')).toHaveTextContent('Value');
  });

  it('applies monospace font when mono is true', () => {
    const { container } = render(<CardRow label="Label" value="12345" mono={true} />);
    const valueEl = container.querySelector('.font-mono');
    expect(valueEl).toBeInTheDocument();
  });

  it('allows wrapping when wrap is true', () => {
    const { container } = render(<CardRow label="Label" value="Long value" wrap={true} />);
    const valueEl = container.querySelector('.break-all');
    expect(valueEl).toBeInTheDocument();
  });

  it('aligns value left when align is left', () => {
    const { container } = render(<CardRow label="Label" value="Value" align="left" />);
    const valueEl = container.querySelector('.text-left');
    expect(valueEl).toBeInTheDocument();
  });

  it('aligns value right by default', () => {
    const { container } = render(<CardRow label="Label" value="Value" />);
    const valueEl = container.querySelector('.text-right');
    expect(valueEl).toBeInTheDocument();
  });
});

describe('CardDivider', () => {
  it('renders an hr element', () => {
    const { container } = render(<CardDivider />);
    expect(container.querySelector('hr')).toBeInTheDocument();
  });

  it('applies custom className', () => {
    const { container } = render(<CardDivider className="custom-divider" />);
    expect(container.querySelector('.custom-divider')).toBeInTheDocument();
  });
});
