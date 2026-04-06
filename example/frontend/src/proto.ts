import protobuf from 'protobufjs/light'
import { registerComponentDecoder } from 'svelgo/runtime/proto'
import appDescriptor from './app_descriptor.json'

const appRoot = protobuf.Root.fromJSON(appDescriptor as protobuf.INamespace)

registerComponentDecoder('Button', appRoot.lookupType('app.ButtonState'))
