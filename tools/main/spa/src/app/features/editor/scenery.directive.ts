import { Directive, ElementRef, Input, OnChanges, inject } from '@angular/core';

declare const BloopScenery: {
  kind(classes: string): string | null;
  catalog: Record<string, { width: number; height: number }>;
  draw(ctx: CanvasRenderingContext2D, kind: string, x: number, y: number, size: number): void;
};

/** Uses the exact gameplay geometry; the underlying grid remains the footprint. */
@Directive({ selector: 'canvas[bloopScenery]', standalone: true })
export class SceneryDirective implements OnChanges {
  @Input() bloopScenery = '';
  @Input() sceneryRow = 0;
  @Input() sceneryThumbnail = false;
  @Input() sceneryBounds = false;
  private readonly element = inject(ElementRef<HTMLCanvasElement>);
  ngOnChanges(): void {
    const canvas = this.element.nativeElement;
    const kind = typeof BloopScenery === 'undefined' ? null : BloopScenery.kind(this.bloopScenery);
    canvas.style.display = kind ? 'block' : 'none';
    if (!kind) return;
    const extent = BloopScenery.catalog[kind];
    const size = this.sceneryThumbnail ? Math.min(90 / extent.width, 90 / extent.height) : 30;
    const width = Math.ceil((extent.width + .4) * size), height = Math.ceil((extent.height + .4) * size);
    canvas.width = width * 2; canvas.height = height * 2;
    Object.assign(canvas.style, { position: 'absolute', width: `${width}px`, height: `${height}px`, bottom: '0', left: `calc(50% - ${width / 2}px)`, pointerEvents: 'none', zIndex: String(100 + this.sceneryRow) });
    const ctx = canvas.getContext('2d');
    if (ctx) { ctx.scale(2,2); BloopScenery.draw(ctx,kind,width/2,height-3,size); if(this.sceneryBounds) {ctx.strokeStyle="#ffdd83";ctx.lineWidth=1;ctx.setLineDash([3,2]);ctx.strokeRect(1,1,width-2,height-2);ctx.setLineDash([]);ctx.fillStyle="#ffe0a0";ctx.beginPath();ctx.arc(width/2,height-3,3,0,Math.PI*2);ctx.fill();} }
    canvas.setAttribute('aria-label',`${kind} scenery; collision is shown on the grid`);
  }
}
