import { registerComponent } from './registry'
import { registerComponentDecoder, root } from './proto'
import Button from '../component/Button.svelte'
import Label from '../component/Label.svelte'

registerComponent('svelgo.Button', Button)
registerComponent('svelgo.Label', Label)
registerComponentDecoder('svelgo.Button', root.lookupType('ui.ButtonState'))
registerComponentDecoder('svelgo.Label', root.lookupType('ui.LabelState'))
