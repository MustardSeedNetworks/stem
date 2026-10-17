/**
 * useFocusTrap had no coverage, which is the wrong state for the code that
 * decides whether a keyboard user can get out of a drawer: a regression here
 * is invisible to a mouse and total to a keyboard.
 *
 * These pin the wrapping behaviour specifically, because the guard that finds
 * the trap's two ends was restructured for noUncheckedIndexedAccess —
 * `focusableElements.length === 0` became a check on the ends themselves. The
 * conditions are equivalent, and equivalence is worth proving rather than
 * asserting.
 */
import { fireEvent } from '@testing-library/dom';
import { render, screen } from '@testing-library/react';
import { useState } from 'react';
import { afterAll, beforeAll, describe, expect, it, vi } from 'vitest';
import { useFocusTrap } from './useFocusTrap';

/**
 * The hook skips candidates whose `offsetParent` is null, so it ignores
 * elements that are not laid out. jsdom implements no layout and reports null
 * for everything, so without this the trap silently finds nothing focusable
 * and every assertion below would pass for the wrong reason — which is how a
 * hook like this ends up with no tests in the first place.
 *
 * The visibility filter itself therefore cannot be tested here; it needs a
 * real browser, and belongs in the Playwright suite.
 */
let realOffsetParent: PropertyDescriptor | undefined;

beforeAll(() => {
  realOffsetParent = Object.getOwnPropertyDescriptor(HTMLElement.prototype, 'offsetParent');
  Object.defineProperty(HTMLElement.prototype, 'offsetParent', {
    configurable: true,
    get(): Element | null {
      return document.body;
    },
  });
});

afterAll(() => {
  if (realOffsetParent) {
    Object.defineProperty(HTMLElement.prototype, 'offsetParent', realOffsetParent);
  }
});

function Trapped({ empty = false }: { empty?: boolean }) {
  const ref = useFocusTrap<HTMLDivElement>({ isActive: true, autoFocus: false });
  return (
    <div ref={ref} data-testid="trap">
      {empty ? null : (
        <>
          <button type="button">first</button>
          <button type="button">middle</button>
          <button type="button">last</button>
        </>
      )}
    </div>
  );
}

function AutoFocused() {
  const ref = useFocusTrap<HTMLDivElement>({ isActive: true });
  return (
    <div ref={ref}>
      <button type="button">close</button>
      <input aria-label="key" />
    </div>
  );
}

/** A drawer whose owner re-renders on demand, with a dialog that opens over it. */
function DialogOverDrawer({
  onDrawerEscape,
  onDialogEscape,
}: {
  onDrawerEscape: () => void;
  onDialogEscape: () => void;
}) {
  const [renders, setRenders] = useState(0);
  const [dialogOpen, setDialogOpen] = useState(false);
  const drawerRef = useFocusTrap<HTMLDivElement>({
    isActive: true,
    autoFocus: false,
    onEscape: () => onDrawerEscape(),
  });
  const dialogRef = useFocusTrap<HTMLDivElement>({
    isActive: dialogOpen,
    autoFocus: false,
    onEscape: () => onDialogEscape(),
  });
  return (
    <>
      <div ref={drawerRef}>
        <button type="button" onClick={() => setDialogOpen(true)}>
          open dialog
        </button>
        <button type="button" onClick={() => setRenders(renders + 1)}>
          re-render {renders}
        </button>
      </div>
      {dialogOpen ? (
        <div ref={dialogRef}>
          <button type="button">in dialog</button>
        </div>
      ) : null}
    </>
  );
}

const nextFrame = (): Promise<void> =>
  new Promise((resolve) => {
    requestAnimationFrame(() => resolve());
  });

describe('useFocusTrap', () => {
  it('focuses the first element once the frame lands', async () => {
    render(<AutoFocused />);

    await nextFrame();

    expect(document.activeElement).toBe(screen.getByRole('button', { name: 'close' }));
  });

  it('leaves focus alone when the user is already inside before the frame (#1372)', async () => {
    render(<AutoFocused />);
    const input = screen.getByRole('textbox', { name: 'key' });
    input.focus();

    await nextFrame();

    expect(document.activeElement).toBe(input);
  });

  it('wraps Tab from the last element back to the first', () => {
    render(<Trapped />);
    screen.getByRole('button', { name: 'last' }).focus();

    fireEvent.keyDown(document, { key: 'Tab' });

    expect(document.activeElement).toBe(screen.getByRole('button', { name: 'first' }));
  });

  it('wraps Shift+Tab from the first element to the last', () => {
    render(<Trapped />);
    screen.getByRole('button', { name: 'first' }).focus();

    fireEvent.keyDown(document, { key: 'Tab', shiftKey: true });

    expect(document.activeElement).toBe(screen.getByRole('button', { name: 'last' }));
  });

  it('leaves Tab alone in the middle of the list', () => {
    render(<Trapped />);
    const middle = screen.getByRole('button', { name: 'middle' });
    middle.focus();

    fireEvent.keyDown(document, { key: 'Tab' });

    /* The browser moves focus itself; the trap only intervenes at the ends. */
    expect(document.activeElement).toBe(middle);
  });

  it('lets only the newest trap act on Escape', () => {
    const onDrawerEscape = vi.fn();
    const onDialogEscape = vi.fn();
    render(<DialogOverDrawer onDrawerEscape={onDrawerEscape} onDialogEscape={onDialogEscape} />);
    fireEvent.click(screen.getByRole('button', { name: 'open dialog' }));

    fireEvent.keyDown(screen.getByRole('button', { name: 'in dialog' }), { key: 'Escape' });

    expect(onDialogEscape).toHaveBeenCalledTimes(1);
    expect(onDrawerEscape).not.toHaveBeenCalled();
  });

  it('keeps the dialog newest when the drawer under it re-renders', () => {
    const onDrawerEscape = vi.fn();
    const onDialogEscape = vi.fn();
    render(<DialogOverDrawer onDrawerEscape={onDrawerEscape} onDialogEscape={onDialogEscape} />);
    fireEvent.click(screen.getByRole('button', { name: 'open dialog' }));
    fireEvent.click(screen.getByRole('button', { name: /^re-render/ }));

    fireEvent.keyDown(screen.getByRole('button', { name: 'in dialog' }), { key: 'Escape' });

    expect(onDialogEscape).toHaveBeenCalledTimes(1);
    expect(onDrawerEscape).not.toHaveBeenCalled();
  });

  it('does nothing at all when the container has nothing focusable', () => {
    render(
      <>
        <button type="button">outside</button>
        <Trapped empty={true} />
      </>,
    );
    const outside = screen.getByRole('button', { name: 'outside' });
    outside.focus();

    /* The empty case is the one the restructured guard covers: with no ends
       to trap between, the handler returns rather than reaching for one. */
    expect(() => fireEvent.keyDown(document, { key: 'Tab' })).not.toThrow();
    expect(document.activeElement).toBe(outside);
  });
});
